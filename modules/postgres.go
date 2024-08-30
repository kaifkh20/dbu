package modules

import (
	"database/sql"
	"fmt"
	"strconv"

	_ "github.com/lib/pq"
)

func ConnectPSQL(config Config) error {
	fmt.Println("Connecting to PSQL..")

	connStr := "user=" + config.User + "password=" + config.Password + "host=" + config.Host + "port=" + strconv.Itoa(config.Port) + "dbname=" + config.Database
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return err
	}

	if pingerr := db.Ping(); pingerr != nil {
		return pingerr
	}

	defer db.Close()

	return nil
}
