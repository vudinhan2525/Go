package main

import (
	"database/sql"
	"log"
	"main/api"
	db "main/db/sqlc"
	"main/pkg/val"
	"main/util"

	_ "github.com/lib/pq"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("Error when loading env!!", err)
	}
	val.RegisterCustomValidations()
	conn, err := sql.Open(config.DbDriver, config.DbSource)
	if err != nil {
		log.Fatal("Error when connecting to db!!", err)
	}
	store := db.NewStore(conn)
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("Error when creating server")
	}
	err = server.StartServer(config.APIEndpoint)
	if err != nil {
		log.Fatal("Error when starting server")
	}
}
