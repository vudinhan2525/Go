package api

import (
	"fmt"
	db "main/db/sqlc"
	"main/pkg/middlewares"
	"main/token"
	"main/util"

	"github.com/gin-gonic/gin"
)

type Server struct {
	Config     util.Config
	TokenMaker token.Maker
	Store      db.Store
	Router     *gin.Engine
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := Server{Store: store, TokenMaker: tokenMaker, Config: config}
	server.SetupRouter()

	return &server, nil
}
func (server *Server) SetupRouter() {
	router := gin.Default()
	router.Use(middlewares.GlobalErrorHandler())
	err := router.SetTrustedProxies(nil)
	if err != nil {
		panic(err)
	}

	router.POST("/login", server.loginUser)
	router.POST("/refresh_token", server.refreshToken)

	privateRouter := router.Group("/").Use(middlewares.AuthMiddleware(server.TokenMaker))

	privateRouter.POST("/users", server.createUser)
	privateRouter.GET("/users/:id", server.getUsertById)

	privateRouter.POST("/accounts", server.createAccount)
	privateRouter.GET("/accounts/:id", server.getAccountById)
	privateRouter.GET("/accounts", server.getAccounts)

	privateRouter.POST("/transfer", server.transferMoney)

	server.Router = router
}
func (server *Server) StartServer(address string) error {
	return server.Router.Run(address)
}
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
