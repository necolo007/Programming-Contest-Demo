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
	//ctx := context.Background()
	//err := milvus.InitMilvus(&ctx)
	//if err != nil {
	//	log.Fatalf("初始化Milvus客户端失败: %v", err)
	//}
	//if !milvus.IsClientInit() {
	//	log.Fatalln("Milvus客户端未初始化")
	//} else {
	//	log.Println("Milvus客户端初始化成功")
	//}
	//err = milvus.DeleteMilvusCollection(&ctx)
	//if err != nil {
	//	log.Fatalf("删除Milvus集合失败: %v", err)
	//}
	//err = milvus.CreateCollection(ctx)
	//if err != nil {
	//	log.Fatalf("创建Milvus集合失败: %v", err)
	//}
	return r
}
