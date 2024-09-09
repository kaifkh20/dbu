package modules

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jamf/go-mysqldump"
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
	fileName := fmt.Sprintf("mysql_backup_%s", timestamp)

	dumper, err := mysqldump.Register(db, outputDir, fileName)

	if err != nil {
		return err
	}

	err = dumper.Dump()

	if err != nil {
		return fmt.Errorf("error backuping : %v", err)
	}

	defer dumper.Close()
	return nil
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
				return fmt.Errorf("error executing sql statements : %v \n %s", err, statement.String())
			}
			statement.Reset()
		}
	}

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("error reading the file specified at the input path %s :%v", inputPath, err)
	}

	return nil
}
