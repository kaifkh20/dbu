package modules

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/go-sql-driver/mysql"
)

// var db *sql.DB

func ConnectMySQL(config Config) error {
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
		return err
	}
	if pingerr := db.Ping(); pingerr != nil {
		return pingerr
	}
	defer db.Close()
	return nil
}
