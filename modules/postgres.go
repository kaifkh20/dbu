package modules

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/JCoupalK/go-pgdump"
	_ "github.com/lib/pq"
)

func ConnectPSQL(config Config) (*sql.DB, error) {
	fmt.Println("Connecting to PSQL..")
	
	// Validate configuration
	if config.User == "" || config.Host == "" || config.Database == "" {
		return nil, fmt.Errorf("missing required configuration: user, host, or database")
	}
	if config.Port <= 0 || config.Port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", config.Port)
	}
	
	// Build connection string with proper escaping
	params := url.Values{}
	params.Set("user", config.User)
	params.Set("password", config.Password)
	params.Set("host", config.Host)
	params.Set("port", strconv.Itoa(config.Port))
	params.Set("dbname", config.Database)
	params.Set("sslmode", "prefer") // Add SSL mode
	params.Set("connect_timeout", "30") // Add timeout
	
	connStr := "postgres://?" + params.Encode()
	
	db, err := sql.Open("postgres", connStr)
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

func BackupPSQL(db *sql.DB, outputDir string, config Config) error {
	// Validate inputs
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
	fileName := fmt.Sprintf("psql_backup_%s.sql", timestamp)
	filePath := filepath.Join(outputDir, fileName)
	
	// Build connection string with proper escaping
	params := url.Values{}
	params.Set("user", config.User)
	params.Set("password", config.Password)
	params.Set("host", config.Host)
	params.Set("port", strconv.Itoa(config.Port))
	params.Set("dbname", config.Database)
	params.Set("sslmode", "prefer")
	
	connStr := "postgres://?" + params.Encode()
	
	fmt.Printf("Creating backup: %s\n", filePath)
	
	// Create backup with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	
	dumper := pgdump.NewDumper(connStr)
	
	done := make(chan error, 1)
	go func() {
		done <- dumper.DumpDatabase(filePath)
	}()
	
	select {
	case err := <-done:
		if err != nil {
			os.Remove(filePath) // Clean up failed backup
			return fmt.Errorf("error backing up database: %w", err)
		}
	case <-ctx.Done():
		os.Remove(filePath) // Clean up on timeout
		return fmt.Errorf("backup timeout: %w", ctx.Err())
	}
	
	stat,err :=  os.Stat(filePath)
	// Verify backup file was created and has content
	if err != nil {
		return fmt.Errorf("backup file verification failed: %w", err)
	} else if stat.Size() == 0 {
		os.Remove(filePath)
		return fmt.Errorf("backup file is empty")
	}
	
	fmt.Printf("Backup completed successfully: %.2f MB\n", float64(stat.Size())/1024/1024)
	return nil
}

func RestorePSQL(db *sql.DB, inputPath string) error {
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
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()
	
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	
	var statement strings.Builder
	var executedStatements int
	
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
