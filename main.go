package main

import (
	"fmt"
	"os"
	dbu "dbu/modules"
	"github.com/akamensky/argparse"
)


func main() {
	parser := argparse.NewParser("dbu", "Backup your database")

	// test := parser.String("p", "print", &argparse.Options{Required: true, Help: "Print your string"})

	nameofdb := parser.Selector("n", "name", []string{"mysql", "postgres", "mongodb"}, &argparse.Options{Required: true, Help: "Specify the DB Provider"})

	host := parser.String("","host",&argparse.Options{Required : true, Help : "Host Name"})

	port := parser.String("p","port",&argparse.Options{Required: true,Help : "Port Number"})
	
	username := parser.String("u","user",&argparse.Options{Required: true,Help:"Specify the User"})

	password := parser.String("","password",&argparse.Options{Required: true,Help:"Specify the Password"})

	database := parser.String("d","database",&argparse.Options{Required : true,Help : "Specify the Database Name"})


	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}

	config := dbu.Config{Host:*host,Port:*port,User:*username,Password:*password,Database:*database,DBProviderName:*nameofdb}
	dbu.InitiateConnection(config)

	
}
