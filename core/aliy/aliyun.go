package aliy

import (
	"Programming-Demo/config"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"log"
)

type Aliyun struct {
	client *sdk.Client
}

var AliyunClient *Aliyun

func InitAliyun() {
	AliyunClient = &Aliyun{}

	conf := sdk.NewConfig()
	conf.AutoRetry = true
	conf.MaxRetryTime = 3
	//连接池配置
	conf.WithGoRoutinePoolSize(10)
	// 创建凭证对象
	credential := credentials.NewStaticAKCredentialsProvider(config.GetConfig().AccessKeyID, config.GetConfig().AccessKeySecret)

	var err error
	AliyunClient.client, err = sdk.NewClientWithOptions(
		"cn-hangzhou",
		conf,
		credential,
	)
	if err != nil {
		log.Println("初始化阿里云客户端失败")
	}
}

func (a *Aliyun) GetClient() *sdk.Client {
	return a.client
}
