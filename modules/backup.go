package modules

import (
	"errors"
	"fmt"
	"log"
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
	Connect() error
	Backup() error
}

func (config Config) Connect() error {
	if config.DBProviderName == "mysql" {
		err := ConnectMySQL(config)
		if err != nil {
			return err
		}
		return nil
	} else if config.DBProviderName == "postgres" {
		err := ConnectPSQL(config)
		if err != nil {
			return err
		}
		return nil
	} else if config.DBProviderName == "mongodb" {
		err := ConnectMongo(config)
		if err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("invalid sql provider")
	}
}

func InitiateConnection(config Config) {
	err := config.Connect()
	if err != nil {
		log.Fatal("Unable to Establish Connection\n", err)
	}
	fmt.Println("Connection Established")
	// for {
	// }
}
