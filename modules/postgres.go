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

	"github.com/JCoupalK/go-pgdump"
	_ "github.com/lib/pq"
)

func ConnectPSQL(config Config) (*sql.DB, error) {
	fmt.Println("Connecting to PSQL..")

	connStr := "user=" + config.User + "password=" + config.Password + "host=" + config.Host + "port=" + strconv.Itoa(config.Port) + "dbname=" + config.Database
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}

	if pingerr := db.Ping(); pingerr != nil {
		return nil, pingerr
	}

	// defer db.Close()

	return db, nil
}

func BackupPSQL(db *sql.DB, outputDir string, config Config) (string,error) {

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	fileName := fmt.Sprintf("psql_backup_%s.sql", timestamp)
	filePath := filepath.Join(outputDir, fileName)

	connStr := "user=" + config.User + "password=" + config.Password + "host=" + config.Host + "port=" + strconv.Itoa(config.Port) + "dbname=" + config.Database
	// db, err := sql.Open("postgres", connStr)

	dumper := pgdump.NewDumper(connStr)

	if err := dumper.DumpDatabase(filePath); err != nil {
		os.Remove(filePath)
		return "",fmt.Errorf("error backing up database : %v", err)
	}

	return filePath,nil
}

func RestorePSQL(db *sql.DB, inputPath string) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	var statement strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "--") || strings.TrimSpace(line) == "" || strings.HasPrefix(line, "/*") || strings.HasSuffix(line, "*/;") {
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
