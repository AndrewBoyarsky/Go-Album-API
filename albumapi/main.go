package main

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/AndrewBoyarsky/albumapi/albums"
	alb "github.com/AndrewBoyarsky/albumapi/albums"
	"github.com/AndrewBoyarsky/albumapi/api"
	"github.com/AndrewBoyarsky/albumapi/db"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var albumsTestData = []alb.Album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

func main() {
	db.ConnectDb()
	StartAPIServer()
}

func StartAPIServer() {
	router := gin.Default()
	gin.CustomRecovery(func(c *gin.Context, err any) {
		logrus.Errorf("Error serving request: %s %s, reason: %s, stacktrace: %w", c.Request.Method, c.Request.RequestURI,
			err, debug.Stack())
		c.AbortWithStatusJSON(http.StatusInternalServerError, api.Error{fmt.Sprintf("Fatal error: %v", err)})
	})
	doAuthGroup := router.Group("/rest/v1")
	doAuthGroup.POST("/register", api.Register)
	doAuthGroup.POST("/login", api.Login)
	authProtectedGroup := router.Group("/rest/v1", api.AuthMiddleware)
	albums.Init(authProtectedGroup)
	err := router.Run("localhost:9123")
	if err != nil {
		logrus.Fatalf("Failed to start API server, maybe 9123 port is busy: %s", err.Error())
	}
}
