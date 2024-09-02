package modules

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

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

func BackupPSQL(db *sql.DB, outputDir string) error {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	fileName := fmt.Sprintf("psql_backup_%s.sql", timestamp)
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
