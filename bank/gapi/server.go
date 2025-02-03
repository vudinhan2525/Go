package gapi

import (
	"fmt"
	db "main/db/sqlc"
	"main/pb"
	"main/token"
	"main/util"
	"main/worker"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
	Config          util.Config
	TokenMaker      token.Maker
	Store           db.Store
	TaskDistributor worker.TaskDistributor
}

func NewServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := Server{Store: store, TokenMaker: tokenMaker, Config: config, TaskDistributor: taskDistributor}
	return &server, nil
}
