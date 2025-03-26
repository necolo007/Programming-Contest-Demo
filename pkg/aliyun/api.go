package aliyun

import (
	"Programming-Demo/core/aliy"
	"encoding/json"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
)

// Response 定义响应结构体
type Response struct {
	RequestId string `json:"RequestId"`
	Data      string `json:"Data"`
}

type InnerData struct {
	Result struct {
		Vec []float64 `json:"vec"`
	} `json:"result"`
	Success bool `json:"success"`
}

func NplApi(s string) []float64 {
	request := requests.NewCommonRequest()
	request.Domain = "alinlp.cn-hangzhou.aliyuncs.com"
	request.Version = "2020-06-29"
	// 因为是RPC接口，因此需指定ApiName(Action)
	request.ApiName = "GetWeChGeneral"
	request.QueryParams["ServiceCode"] = "alinlp"
	request.QueryParams["Text"] = s
	request.QueryParams["TokenizerId"] = "GENERAL_CHN"
	response, err := aliy.AliyunClient.GetClient().ProcessCommonRequest(request)
	if err != nil {
		panic(err)
	}
	// 解析外层响应
	var resp Response
	if err := json.Unmarshal([]byte(response.GetHttpContentString()), &resp); err != nil {
		panic(err)
	}

	// 解析内层数据
	var innerData InnerData
	if err := json.Unmarshal([]byte(resp.Data), &innerData); err != nil {
		panic(err)
	}
	return innerData.Result.Vec
}
