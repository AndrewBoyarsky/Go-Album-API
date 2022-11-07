package albums

import "github.com/gin-gonic/gin"

var repo AlbumRepo

func Init(group *gin.RouterGroup) {
	repo = NewAlbumRepo()
	registerEndpoints(group)
}
