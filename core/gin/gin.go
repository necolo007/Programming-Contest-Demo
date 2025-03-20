package gin

import (
	"Programming-Demo/core/auth"
	"Programming-Demo/core/client"
	"Programming-Demo/core/gin/dbs"
	"Programming-Demo/core/milvus"
	"Programming-Demo/internal/router"
	"github.com/gin-gonic/gin"
	"log"
)

func GinInit() *gin.Engine {
	r := gin.Default()
	dbs.InitDB()
	client.InitClient()
	router.GenerateRouters(r)

	auth.InitSecret()
	err := milvus.InitMilvus()
	if err != nil {
		log.Fatalf("初始化Milvus客户端失败: %v", err)
	}
	if !milvus.IsClientInit() {
		log.Fatalln("Milvus客户端未初始化")
	} else {
		log.Println("Milvus客户端初始化成功")
	}
	return r
}
