package milvus

import (
	"Programming-Demo/config"
	"context"
	"fmt"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

type Client struct {
	client *client.Client
}

var MilvusClient *Client

func (c *Client) GetClient() client.Client {
	return *c.client
}

// InitMilvus 初始化Milvus客户端
func InitMilvus() error {
	ctx := context.Background()
	cfg := config.GetConfig()

	MilvusClient = &Client{}
	c, err := client.NewGrpcClient(ctx, fmt.Sprintf("%s:%d", cfg.Milvus.Host, cfg.Milvus.Port))
	if err != nil {
		return err
	}
	MilvusClient.client = &c
	// 创建集合
	err = createCollection(ctx)
	if err != nil {
		return err
	}

	return nil
}

func IsClientInit() bool {
	return MilvusClient != nil && MilvusClient.client != nil
}

// 创建集合
func createCollection(ctx context.Context) error {
	collName := config.GetConfig().Milvus.Collection
	dim := int64(config.GetConfig().Milvus.Dim)

	schema := &entity.Schema{
		CollectionName: collName,
		Description:    "法律文档向量集合",
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeInt64,
				PrimaryKey: true,
				AutoID:     true,
			},
			{
				Name:     "content",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					"max_length": "65535",
				},
			},
			{
				Name:     "vector",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": fmt.Sprintf("%d", dim),
				},
			},
		},
	}

	exist, err := MilvusClient.GetClient().HasCollection(ctx, collName)
	if err != nil {
		return err
	}
	if !exist {
		err = MilvusClient.GetClient().CreateCollection(ctx, schema, 1)
		if err != nil {
			return err
		}
	}

	return nil
}

func InsertVectors(vectors [][]float32, contents []string) error {
	if !IsClientInit() {
		return fmt.Errorf("milvus 客户端未正确初始化")
	}

	ctx := context.Background()
	collName := config.GetConfig().Milvus.Collection

	// 只创建集合中定义的两个字段
	contentColumn := entity.NewColumnVarChar("content", contents)
	dim := config.GetConfig().Milvus.Dim
	vectorColumn := entity.NewColumnFloatVector("vector", dim, vectors)

	// 只传递两个字段
	_, err := MilvusClient.GetClient().Insert(
		ctx,
		collName,
		"",
		contentColumn, vectorColumn,
	)

	if err != nil {
		return fmt.Errorf("插入数据失败: %v", err)
	}

	return nil
}

// SearchVectors 搜索相似向量并返回对应内容
func SearchVectors(vector []float32, topK int) ([]int64, []float32, []string, error) {
	ctx := context.Background()
	collName := config.GetConfig().Milvus.Collection

	// 创建搜索参数
	sp, err := entity.NewIndexIvfFlatSearchParam(10)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("创建搜索参数失败: %v", err)
	}

	// 创建搜索向量
	searchVector := entity.FloatVector(vector)
	searchVectors := []entity.Vector{searchVector}

	// 执行搜索
	results, err := MilvusClient.GetClient().Search(
		ctx,                           // 上下文
		collName,                      // 集合名称
		[]string{},                    // 分区名称，空字符串表示全部分区
		"",                            // 表达式，用于筛选
		[]string{"content", "vector"}, // 输出字段
		searchVectors,                 // 搜索向量
		"vector",                      // 向量字段名称
		entity.L2,                     // 度量类型
		int(int64(topK)),              // 搜索结果数量
		sp,                            // 搜索参数
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("搜索向量失败: %v", err)
	}

	if len(results) == 0 {
		return nil, nil, nil, fmt.Errorf("未找到匹配结果")
	}

	// 处理搜索结果
	ids := make([]int64, 0, topK)
	scores := make([]float32, 0, topK)
	contents := make([]string, 0, topK)

	for _, result := range results {
		// 提取ID
		if result.IDs != nil {
			idCol, ok := result.IDs.(*entity.ColumnInt64)
			if !ok {
				return nil, nil, nil, fmt.Errorf("无效的ID类型")
			}
			ids = append(ids, idCol.Data()...)
			scores = append(scores, result.Scores...)
		}

		// 提取content
		for _, field := range result.Fields {
			if contentCol, ok := field.(*entity.ColumnVarChar); ok {
				contents = append(contents, contentCol.Data()...)
			}
		}
	}

	return ids, scores, contents, nil
}
