package modules

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

// var db *sql.DB

func ConnectMySQL(config Config) (*sql.DB, error) {
	fmt.Println("Connecting to MYSQL...")
	cfg := mysql.Config{
		User:   config.User,
		Passwd: config.Password,
		Net:    "tcp",
		Addr:   config.Host + ":" + strconv.Itoa(config.Port),
		DBName: config.Database,
	}
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}
	if pingerr := db.Ping(); pingerr != nil {
		return nil, pingerr
	}
	// defer db.Close()
	return db, nil
}

func BackupMYSQL(db *sql.DB, outputDir string) error {

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	fileName := fmt.Sprintf("mysql_backup_%s.sql", timestamp)
	filePath := filepath.Join(outputDir, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating backup file :%v", err)
	}
	defer file.Close()

	tables, err := getTable(db)
	if err != nil {
		return err
	}

	for _, table := range tables {
		if err := dumpTable(db, file, table); err != nil {
			return fmt.Errorf("error dumping table %s: %v", table, err)
		}
	}

	return nil
}

func getTable(db *sql.DB) ([]string, error) {

	var tables []string

	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	return tables, nil
}

func dumpTable(db *sql.DB, file *os.File, table string) error {

	var tableCreate string

	err := db.QueryRow("SHOW CREATE TABLE "+table).Scan(&table, &tableCreate)

	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(file, "%s\n\n", tableCreate)

	if err != nil {
		return err
	}

	rows, err := db.Query("SELECT * FROM " + table)
	if err != nil {
		return err
	}

	defer rows.Close()

	columns, err := rows.Columns()

	if err != nil {
		return err
	}

	values := make(sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))

	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			return err
		}
		var valueStr []string

		for _, col := range values {
			if col == 0 {
				valueStr = append(valueStr, "NULL")
			} else {
				valueStr = append(valueStr, fmt.Sprintf("'%s'", escapeString(string(col))))
			}

		}

		_, err = fmt.Fprintf(file, "INSERT INTO %s (%s) VALUES (%s);\n", table, strings.Join(columns, ","), strings.Join(valueStr, ","))
		if err != nil {
			return err
		}
	}

	return nil
}

func escapeString(s string) string {
	return strings.NewReplacer("'", "''", "\\", "\\\\").Replace(s)
}

func RestoreMYSQL(db *sql.DB, inputPath string) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	var statement strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix("line", "--") || strings.TrimSpace(line) == "" {
			continue
		}

		statement.WriteString(line)
		statement.WriteString(" ")

		if strings.HasSuffix(strings.TrimSpace(line), ";") {
			_, err := db.Exec(statement.String())
			if err != nil {
				return fmt.Errorf("Error executing SQL Statements : %v \n %s", err, statement.String())
			}
			statement.Reset()
		}
	}

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("Error Reading The File specified at the input path %s :%v", inputPath, err)
	}

	return nil
}
