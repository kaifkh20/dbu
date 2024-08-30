package main

import (
	dbu "dbu/modules"
	"log"
	"os"
	"strconv"

	"github.com/akamensky/argparse"
)

func main() {
	parser := argparse.NewParser("dbu", "Backup your database")

	nameofdb := parser.Selector("n", "name", []string{"mysql", "postgres", "mongodb"}, &argparse.Options{Required: true, Help: "Specify the DB Provider"})

	host := parser.String("", "host", &argparse.Options{Required: true, Help: "Host Name"})

	port := parser.String("p", "port", &argparse.Options{Required: true, Help: "Port Number"})

	username := parser.String("u", "user", &argparse.Options{Required: true, Help: "Specify the User"})

	password := parser.String("", "password", &argparse.Options{Required: true, Help: "Specify the Password"})

	database := parser.String("d", "database", &argparse.Options{Required: true, Help: "Specify the Database Name"})

	err := parser.Parse(os.Args)
	if err != nil {
		log.Fatal(parser.Usage(err))
	}

	portno, err := strconv.Atoi(*port)

	if err != nil {
		log.Fatal(err)
	}

	config := dbu.Config{Host: *host, Port: portno, User: *username, Password: *password, Database: *database, DBProviderName: *nameofdb}
	dbu.InitiateConnection(config)

}
