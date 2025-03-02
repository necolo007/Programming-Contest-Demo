package gin

import (
	"Programming-Demo/core/auth"
	"Programming-Demo/core/client"
	"Programming-Demo/core/gin/dbs"
	"Programming-Demo/internal/router"
	"github.com/gin-gonic/gin"
)

func GinInit() *gin.Engine {
	r := gin.Default()
	dbs.InitDB()
	client.InitClient()
	router.GenerateRouters(r)

	auth.InitSecret()
	return r
}
