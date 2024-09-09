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
	Restore(*sql.DB, string) error
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
		err := BackupPSQL(db, outputDir, config)
		if err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("invalid sql provider")
	}
}

func (config Config) Restore(db *sql.DB, inputPath string) error {

	if config.DBProviderName == "mysql" {
		err := RestoreMYSQL(db, inputPath)
		if err != nil {
			return err
		}
		return nil
	} else if config.DBProviderName == "postgres" {
		err := RestorePSQL(db, inputPath)
		if err != nil {
			return err
		}
		return nil
	}

	return nil
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
		// fmt.Scanln()
		// var err error
		if choice > 3 {
			fmt.Println("Invalid choice.")
		}

		if choice == 1 {
			fmt.Print("Specify the path directory: ")
			var outputDir string
			// Corrected input handling
			_, err := fmt.Scanln(&outputDir) // Removed the format specifier
			if err != nil {
				log.Fatal(err) // Handle any input errors
			}

			err = config.Backup(db, outputDir)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Backup done...")
		} else if choice == 2 {
			fmt.Print("Specify the path directory of the backup file: ")
			var inputPath string
			// Corrected input handling
			_, err := fmt.Scanln(&inputPath) // Removed the format specifier
			if err != nil {
				log.Fatal(err) // Handle any input errors
			}

			fmt.Println("Directory:", inputPath)
			err = config.Restore(db, inputPath)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Restoration done.")
		} else {
			os.Exit(0)
		}
	}
}
