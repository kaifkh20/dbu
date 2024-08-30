package modules

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func ConnectMongo(config Config) error {

	fmt.Println("Connecting to MonogoDB...")

	curi := "mongodb://+" + "[" + config.User + ":" + config.Password + "]" + "@" + config.Host + ":" + strconv.Itoa(config.Port) + "/" + "[" + config.Database + "]"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(curi))
	if err != nil {
		// cancel()
		return err
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		// cancel()
		return err
	}

	defer client.Disconnect(ctx)

	return nil
}
