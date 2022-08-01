package main

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/leandro-koller-bft/video-encoder/application/services"
	"github.com/leandro-koller-bft/video-encoder/framework/database"
	"github.com/leandro-koller-bft/video-encoder/framework/queue"
	"github.com/leandro-koller-bft/video-encoder/local_constants"
	"github.com/streadway/amqp"
)

var db database.Database

func init() {
	err := godotenv.Load()
	throwFatalError(err, "Error while loading .env file")
	autoMigrateDB, err := strconv.ParseBool(os.Getenv(local_constants.AUTO_MIGRATE_DB_ENV))
	throwFatalError(err, "Error while parsing boolean env var (AUTO_MIGRATE_DB)")
	debug, err := strconv.ParseBool(os.Getenv(local_constants.DEBUG_ENV))
	throwFatalError(err, "Error while parsing boolean env var (DEBUG)")

	db.AutoMigrateDB = autoMigrateDB
	db.Debug = debug
	db.DSNTest = os.Getenv(local_constants.DSN_TEST_ENV)
	db.DSN = os.Getenv(local_constants.DSN_ENV)
	db.DBTypeTest = os.Getenv(local_constants.DB_TYPE_TEST_ENV)
	db.DBType = os.Getenv(local_constants.DB_TYPE_ENV)
	db.Env = os.Getenv(local_constants.ENV_ENV)
}

func throwFatalError(err error, msg string) {
	if err != nil {
		log.Fatalf(msg)
	}
}

func main() {
	messageChannel := make(chan amqp.Delivery)
	jobReturnChannel := make(chan services.JobWorkerResult)
	dbConnection, err := db.Connect()

	throwFatalError(err, "error connecting to DB")
	defer dbConnection.Close()

	rabbitMQ := queue.NewRabbitMQ()
	ch := rabbitMQ.Connect()
	defer ch.Close()

	rabbitMQ.Consume(messageChannel)

	jobManager := services.NewJobManager(dbConnection, rabbitMQ, jobReturnChannel, messageChannel)
	jobManager.Start(ch)
}
