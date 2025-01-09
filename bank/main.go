package main

import (
	"database/sql"
	"log"
	"main/api"
	db "main/db/sqlc"
	"main/util"

	_ "github.com/lib/pq"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("Error when loading env!!", err)
	}
	conn, err := sql.Open(config.DbDriver, config.DbSource)
	if err != nil {
		log.Fatal("Error when connecting to db!!", err)
	}
	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.StartServer(config.APIEndpoint)
	if err != nil {
		log.Fatal("Error when starting server")
	}
}
