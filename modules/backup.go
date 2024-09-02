package modules

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
)

type Config struct {
	Host           string
	Port           int
	User           string
	Password       string
	Database       string
	DBProviderName string
}

type Database interface {
	Connect() (*sql.DB, error)
	Backup(*sql.DB, string) error
}

func (config Config) Connect() (*sql.DB, error) {
	if config.DBProviderName == "mysql" {
		db, err := ConnectMySQL(config)
		if err != nil {
			return nil, err
		}
		return db, nil
	} else if config.DBProviderName == "postgres" {
		db, err := ConnectPSQL(config)
		if err != nil {
			return nil, err
		}
		return db, nil
	} else {
		return nil, errors.New("invalid sql provider")
	}
}

func (config Config) Backup(db *sql.DB, outputDir string) error {
	if config.DBProviderName == "mysql" {
		err := BackupMYSQL(db, outputDir)
		if err != nil {
			return err
		}
		return nil
	} else if config.DBProviderName == "postgres" {
		err := BackupPSQL(db, outputDir)
		if err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("invalid sql provider")
	}
}

func InitiateConnection(config Config) {
	db, err := config.Connect()
	if err != nil {
		log.Fatal("Unable to Establish Connection\n", err)
	}
	fmt.Println("Connection Established")
	for {
		fmt.Println("1) Backup\n2) Restore(Under-Development)\n3)Type 'exit' to exit.")
		var choice int
		fmt.Scanf("%d", &choice)
		var err error
		if choice > 3 {
			fmt.Println("Invalid choice.")
		}
		if choice == 1 {
			fmt.Println("Specify the path directory.")
			var outputDir string
			fmt.Scanf("%s", outputDir)
			err = config.Backup(db, outputDir)
		} else if choice == 2 {
			fmt.Println("Specify the path directory.")
			var outputDir string
			fmt.Scanf("%s", outputDir)
			err = config.Backup(db, outputDir)

		} else {
			os.Exit(0)
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Backup done...")
	}
}
