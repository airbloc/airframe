package apiserver

import (
	"fmt"
	"github.com/airbloc/airframe/database"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Server struct {
	router *gin.Engine
	port   string
}

func NewServer(backend database.Database, port int) *Server {
	r := gin.Default()
	RegisterV1API(r, backend)
	r.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"foo": "bar"})
	})
	return &Server{
		router: r,
		port:   fmt.Sprintf(":%d", port),
	}
}

func (s *Server) Start() error {
	return s.router.Run(s.port)
}
