package main

import (
	"database/sql"
	"log"
	"main/api"
	db "main/db/sqlc"
	"main/gapi"
	"main/pb"
	"main/pkg/val"
	"main/util"
	"net"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
	runGrpcServer(config, store)
}
func runHttpServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("Error when creating server")
	}
	err = server.StartServer(config.APIEndpoint)
	if err != nil {
		log.Fatal("Error when starting server")
	}
}
func runGrpcServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("Error when creating server")
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GrpcAPIEndpoint)
	if err != nil {
		log.Fatal("Error when creating listener")
	}
	log.Printf("stat gRPC server at %s", listener.Addr().String())

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("Cannot creating grpc server")
	}
}
