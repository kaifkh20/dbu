package modules

import ( "fmt"
	"errors"
	"log"
)

type Config struct {
	Host string
	Port string
	User string
	Password string
	Database string
	DBProviderName string
}



type Database interface {
	Connect() error
	Backup() error
}

func (config Config) Connect() error{
	if config.DBProviderName == "mysql"{
		err := ConnectMySQL(config);
		if err!=nil{
			return err;
		} 
		return nil;
	} else if config.DBProviderName == "postgres" {
		err:= ConnectPSQL(config);
		if err!=nil{
			return err;
		} 
		return nil;
	} else if config.DBProviderName == "mongodb"{
		err:= ConnectMongo(config);
		if err!=nil{
			return err;
		} 
		return nil;
	} else{
		return errors.New("Invalid SQL Provider"); 
	}
}

func InitiateConnection(config Config){
	err := config.Connect();
	if err!=nil{
		log.Fatal("Unable to Establish Connection\n",err);
	}
	fmt.Println("Connection Established");
	for {
	}
}

