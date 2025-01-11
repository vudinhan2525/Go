package api

import (
	db "main/db/sqlc"

	"github.com/gin-gonic/gin"
)

type Server struct {
	store  db.Store
	Router *gin.Engine
}

func NewServer(store db.Store) *Server {
	server := Server{store: store}
	router := gin.Default()
	err := router.SetTrustedProxies(nil)
	if err != nil {
		panic(err)
	}
	router.POST("/users", server.createUser)
	router.GET("/users/:id", server.getUsertById)

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccountById)
	router.GET("/accounts", server.getAccounts)

	router.POST("/transfer", server.transferMoney)

	server.Router = router
	return &server
}

func (server *Server) StartServer(address string) error {
	return server.Router.Run(address)
}
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
