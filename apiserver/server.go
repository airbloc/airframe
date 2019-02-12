package apiserver

import (
	"fmt"
	"github.com/airbloc/airframe/database"
	"github.com/airbloc/logger"
	"github.com/gin-gonic/gin"
	"net/http"
)

var (
	log = logger.New("apiserver")
)

type Server struct {
	server *http.Server
}

func New(backend database.Database, port int, debug bool) *Server {
	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(Logger())
	r.Use(Recovery())
	r.NoRoute(NotFound())

	RegisterV1API(r, backend)

	return &Server{
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: r,
		},
	}
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		timer := log.Timer()
		c.Next()
		timer.End("HTTP", logger.Attrs{
			"method": c.Request.Method,
			"url":    getRequestPath(c.Request),
			"status": c.Writer.Status(),
			"client": c.ClientIP(),
		})
	}
}

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("panic recovered: %v", err, logger.Attrs{
					"method": c.Request.Method,
					"url":    getRequestPath(c.Request),
					"client": c.ClientIP(),
				})
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
		}()
		c.Next()
	}
}

func NotFound() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	}
}

func getRequestPath(r *http.Request) string {
	path := r.URL.Path
	raw := r.URL.RawQuery
	if raw != "" {
		return path + "?" + raw
	}
	return path
}

func (s *Server) Start() error {
	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Stop() {
	if err := s.server.Close(); err != nil {
		log.Error("failed to close HTTP server", err)
	}
}
