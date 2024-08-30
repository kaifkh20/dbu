package main

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
)

func main() {
	parser := argparse.NewParser("dbu", "Backup your database")

	test := parser.String("p", "print", &argparse.Options{Required: true, Help: "Print your string"})

	nameofdb := parser.Selector("n", "name", []string{"mysql", "postgres", "mongodb"}, &argparse.Options{Required: true, Help: "Specify the DB Provider"})

	if *nameofdb == "mysql" {
		connection
	}

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}

	fmt.Println(*test)
}
