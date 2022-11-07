package albums

import (
	"context"
	"net/http"

	"github.com/AndrewBoyarsky/albumapi/api"
	"github.com/AndrewBoyarsky/albumapi/db"
	"github.com/AndrewBoyarsky/albumapi/users"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func getAlbums(c *gin.Context) {
	user := ensureLoggedIn(c)
	albums := repo.GetAll(nil, user.UserName)
	c.IndentedJSON(http.StatusOK, albums)
}

func addNewAlbum(c *gin.Context) {
	newAlbum := Album{}
	if err := c.ShouldBindJSON(&newAlbum); err != nil {
		c.AbortWithStatusJSON(400, api.Error{Error: "bad Album json: " + err.Error()})
	} else {
		user := ensureLoggedIn(c)
		session, _ := db.MongoClient.StartSession()
		defer session.EndSession(context.Background())
		ctx := context.Background()
		id, _ := session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
			newAlbum.UserName = user.UserName
			generatedId := repo.Save(sessCtx, newAlbum, "")
			produceToKafka(&newAlbum, generatedId)
			return generatedId, nil
		})

		c.IndentedJSON(http.StatusCreated, struct {
			Id string
			// }{generatedId})
		}{(id.(string))})
	}
}

func updateAlbum(c *gin.Context) {
	newAlbum := Album{}
	param := c.Param("id")
	if err := c.ShouldBindJSON(&newAlbum); err != nil {
		c.AbortWithStatusJSON(400, api.Error{Error: "bad Album json: " + err.Error()})
	} else {
		user := ensureLoggedIn(c)
		newAlbum.UserName = user.UserName
		savedId := repo.Save(nil, newAlbum, param)
		if savedId == "" {
			c.AbortWithStatusJSON(404, api.Error{Error: "Album with id " + param + " was not found"})
		} else {
			c.Status(http.StatusNoContent)
		}
	}
}

func getAlbumById(c *gin.Context) {
	id := c.Param("id")
	user := ensureLoggedIn(c)
	album := repo.GetById(nil, id, user.UserName)
	if album == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, api.Error{Error: "Album by id was not found: " + id})
	} else {
		c.IndentedJSON(http.StatusOK, album)
	}
}

func deleteById(c *gin.Context) {
	id := c.Param("id")
	user := ensureLoggedIn(c)
	deleted := repo.DeleteById(nil, id, user.UserName)
	if !deleted {
		c.AbortWithStatusJSON(http.StatusNotFound, api.Error{Error: "Album by id was not found: " + id})
	} else {
		c.Status(http.StatusNoContent)
	}
}

func ensureLoggedIn(c *gin.Context) *users.User {
	user, exists := c.Get("user")
	if !exists {
		panic("Fatal error, unknown user accessed album endpoint")
	}
	castedUser := user.(*users.User)
	return castedUser
}

func registerEndpoints(group *gin.RouterGroup) {
	group.GET("/albums", getAlbums)
	group.GET("/albums/:id", getAlbumById)
	group.DELETE("/albums/:id", deleteById)
	group.POST("/albums", addNewAlbum)
	group.PUT("/albums/:id", updateAlbum)
}
