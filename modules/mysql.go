package modules

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/go-sql-driver/mysql"
	"github.com/jamf/go-mysqldump"
)

func ConnectMySQL(config Config) (*sql.DB, error) {
	fmt.Println("Connecting to MYSQL...")
	
	// Validate configuration
	if config.User == "" || config.Host == "" || config.Database == "" {
		return nil, fmt.Errorf("missing required configuration: user, host, or database")
	}
	if config.Port <= 0 || config.Port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", config.Port)
	}
	
	cfg := mysql.Config{
		User:                 config.User,
		Passwd:               config.Password,
		Net:                  "tcp",
		Addr:                 config.Host + ":" + strconv.Itoa(config.Port),
		DBName:               config.Database,
		ParseTime:            true,  // Enable time parsing
		Loc:                  time.UTC, // Set timezone
		Timeout:              30 * time.Second, // Connection timeout
		ReadTimeout:          30 * time.Second, // Read timeout
		WriteTimeout:         30 * time.Second, // Write timeout
		AllowNativePasswords: true, // Support native passwords
	}
	
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}
	
	// Configure connection pool for better performance
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	
	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	return db, nil
}

func BackupMYSQL(db *sql.DB, outputDir string) error {
	if outputDir == "" {
		return fmt.Errorf("output directory is required")
	}
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}
	
	// Ensure output directory exists and is writable
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	// Test write permissions
	testFile := filepath.Join(outputDir, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("output directory is not writable: %w", err)
	}
	os.Remove(testFile)
	
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	fileName := fmt.Sprintf("mysql_backup_%s", timestamp)
	
	fmt.Printf("Creating backup: %s\n", filepath.Join(outputDir, fileName))
	
	// Create backup with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	
	done := make(chan error, 1)
	go func() {
		dumper, err := mysqldump.Register(db, outputDir, fileName)
		if err != nil {
			done <- fmt.Errorf("failed to register mysqldump: %w", err)
			return
		}
		defer dumper.Close()
		
		err = dumper.Dump()
		if err != nil {
			done <- fmt.Errorf("error backing up: %w", err)
			return
		}
		done <- nil
	}()
	
	select {
	case err := <-done:
		if err != nil {
			pattern := filepath.Join(outputDir, fileName+"*")
			if matches, _ := filepath.Glob(pattern); matches != nil {
				for _, match := range matches {
					os.Remove(match)
				}
			}
			return err
		}
	case <-ctx.Done():
		pattern := filepath.Join(outputDir, fileName+"*")
		if matches, _ := filepath.Glob(pattern); matches != nil {
			for _, match := range matches {
				os.Remove(match)
			}
		}
		return fmt.Errorf("backup timeout: %w", ctx.Err())
	}
	
	pattern := filepath.Join(outputDir, fileName+"*")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return fmt.Errorf("backup verification failed: no backup files found")
	}
	
	// Calculate total backup size
	var totalSize int64
	for _, match := range matches {
		if stat, err := os.Stat(match); err == nil {
			totalSize += stat.Size()
		}
	}
	
	fmt.Printf("Backup completed successfully: %.2f MB (%d files)\n", 
		float64(totalSize)/1024/1024, len(matches))
	return nil
}

func RestoreMYSQL(db *sql.DB, inputPath string) error {
	// Validate inputs
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}
	if inputPath == "" {
		return fmt.Errorf("input path is required")
	}
	
	// Check if file exists and is readable
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()
	
	// Verify file has content
	if stat, err := file.Stat(); err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	} else if stat.Size() == 0 {
		return fmt.Errorf("backup file is empty")
	}
	
	fmt.Printf("Starting restore from: %s\n", inputPath)
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()
	
	scanner := bufio.NewScanner(file)
	// Increase buffer size for large SQL statements
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	
	var statement strings.Builder
	var executedStatements int
	
	// Improved comment detection
	singleLineComment := regexp.MustCompile(`^\s*--`)
	multiLineCommentStart := regexp.MustCompile(`/\*`)
	multiLineCommentEnd := regexp.MustCompile(`\*/`)
	inMultiLineComment := false
	
	for scanner.Scan() {
		// Check for timeout
		select {
		case <-ctx.Done():
			return fmt.Errorf("restore timeout: %w", ctx.Err())
		default:
		}
		
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		
		// Skip empty lines
		if trimmedLine == "" {
			continue
		}
		
		// Handle multi-line comments
		if multiLineCommentStart.MatchString(line) {
			inMultiLineComment = true
		}
		if inMultiLineComment {
			if multiLineCommentEnd.MatchString(line) {
				inMultiLineComment = false
			}
			continue
		}
		
		// Skip single-line comments and old style multi-line comments
		if singleLineComment.MatchString(line) || 
		   strings.HasSuffix(trimmedLine, "*/;") {
			continue
		}
		
		statement.WriteString(line)
		statement.WriteString(" ")
		
		if strings.HasSuffix(trimmedLine, ";") {
			sqlStmt := strings.TrimSpace(statement.String())
			if sqlStmt != "" && sqlStmt != ";" {
				_, err := db.ExecContext(ctx, sqlStmt)
				if err != nil {
					return fmt.Errorf("error executing sql statement %d: %w\nSQL: %s", 
						executedStatements+1, err, sqlStmt)
				}
				executedStatements++
				
				// Progress indicator
				if executedStatements%100 == 0 {
					fmt.Printf("Processed %d statements...\n", executedStatements)
				}
			}
			statement.Reset()
		}
	}
	
	if err = scanner.Err(); err != nil {
		return fmt.Errorf("error reading the file specified at the input path %s: %w", inputPath, err)
	}
	
	fmt.Printf("Restore completed successfully. Executed %d statements.\n", executedStatements)
	return nil
}
