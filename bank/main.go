package main

import (
	"context"
	"database/sql"
	"main/api"
	db "main/db/sqlc"
	"main/gapi"
	"main/pb"
	"main/pkg/interceptors"
	"main/pkg/log"
	pkg "main/pkg/mail"
	"main/pkg/val"
	"main/util"
	"main/worker"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hibiken/asynq"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Logger.Fatal("Error when loading env!!", err)
	}
	val.RegisterCustomValidations()
	conn, err := sql.Open(config.DbDriver, config.DbSource)
	if err != nil {
		log.Logger.Fatal("Error when connecting to db!!", err)
	}
	store := db.NewStore(conn)

	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)
	go runGrpcServer(config, store, taskDistributor)
	go runTaskProcessor(config, redisOpt, store)
	runGatewayServer(config, store, taskDistributor)

	//runHttpServer(config, store)
}
func runHttpServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Logger.Fatal("Error when creating server")
	}
	err = server.StartServer(config.APIEndpoint)
	if err != nil {
		log.Logger.Fatal("Error when starting server")
	}
}
func runGrpcServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Logger.Fatal("Error when creating server")
	}

	interceptor := interceptors.NewGRPCInterceptor(server.TokenMaker)
	grpcServer := grpc.NewServer(interceptor.Unary())
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GrpcAPIEndpoint)
	if err != nil {
		log.Logger.Fatal("Error when creating listener")
	}
	log.Logger.Printf("start gRPC server at %s", listener.Addr().String())

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Logger.Fatal("Cannot creating grpc server")
	}
}
func runGatewayServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Logger.Fatal("Error when creating server")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	grpcMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}))

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Logger.Fatal("Error when creating gateway server")
		return
	}
	interceptor := interceptors.NewGatewayInterceptor(server.TokenMaker)

	mux := http.NewServeMux()
	wrappedHandler := interceptor.LoggerMiddleware(interceptor.AuthMiddleware(ctx, grpcMux))

	mux.Handle("/", wrappedHandler)

	listener, err := net.Listen("tcp", config.APIEndpoint)
	if err != nil {
		log.Logger.Fatal("Error when creating listener")
	}
	log.Logger.Printf("start gateway server at %s", listener.Addr().String())

	err = http.Serve(listener, mux)
	if err != nil {
		log.Logger.Fatal("Cannot creating gateway server")
	}
}
func runTaskProcessor(config util.Config, redisOpt asynq.RedisClientOpt, store db.Store) {
	mailer := pkg.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, mailer)
	log.Logger.Printf("start task processor")
	err := taskProcessor.Start()
	if err != nil {
		log.Logger.Fatal("error when start task processor")
	}
}
