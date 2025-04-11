package gin

import (
	bochalient "Programming-Demo/core/Bocha_client"
	"Programming-Demo/core/aliy"
	"Programming-Demo/core/auth"
	"Programming-Demo/core/client"
	"Programming-Demo/core/gin/dbs"
	"Programming-Demo/core/milvus"
	"Programming-Demo/internal/router"
	"context"
	"github.com/gin-gonic/gin"
	"log"
)

func GinInit() *gin.Engine {
	r := gin.Default()
	dbs.InitDB()
	client.InitClient()
	bochalient.InitBochaClient()
	router.GenerateRouters(r)

	auth.InitSecret()
	// 初始化阿里云客户端
	aliy.InitAliyun()
	// 初始化Milvus客户端
	ctx := context.Background()
	err := milvus.InitMilvus(&ctx)
	if err != nil {
		log.Fatalf("初始化Milvus客户端失败: %v", err)
	}
	if !milvus.IsClientInit() {
		log.Fatalln("Milvus客户端未初始化")
	} else {
		log.Println("Milvus客户端初始化成功")
	}
	//err = milvus.DeleteMilvusCollection(&ctx)
	//if err != nil {
	//	log.Printf("删除Milvus集合失败: %v", err)
	//}
	//err = milvus.CreateCollection(ctx)
	//if err != nil {
	//	log.Printf("创建Milvus集合失败: %v", err)
	//}
	return r
}
