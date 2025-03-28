package milvus

import (
	"Programming-Demo/config"
	"context"
	"fmt"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"log"
)

type Client struct {
	client *client.Client
}

var MilvusClient *Client

func (c *Client) GetClient() client.Client {
	return *c.client
}

// InitMilvus 初始化Milvus客户端
func InitMilvus(ctx *context.Context) error {
	cfg := config.GetConfig()

	MilvusClient = &Client{}
	c, err := client.NewGrpcClient(*ctx, fmt.Sprintf("%s:%d", cfg.Milvus.Host, cfg.Milvus.Port))
	if err != nil {
		return err
	}
	MilvusClient.client = &c

	return nil
}

func IsClientInit() bool {
	return MilvusClient != nil && MilvusClient.client != nil
}

// CreateCollection 创建集合
func CreateCollection(ctx context.Context) error {
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
		log.Printf("成功创建集合: %s", collName)
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
	if err := LoadCollection(ctx); err != nil {
		return nil, nil, nil, fmt.Errorf("加载集合失败: %v", err)
	}
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

// SearchVectorsWithParams 添加高级搜索参数的向量搜索
func SearchVectorsWithParams(vector []float32, topK int, params map[string]interface{}) ([]int64, []float32, []string, error) {
	ctx := context.Background()
	if err := LoadCollection(ctx); err != nil {
		return nil, nil, nil, fmt.Errorf("加载集合失败: %v", err)
	}
	collName := config.GetConfig().Milvus.Collection

	// 提取搜索参数
	metricType := entity.L2
	if mt, ok := params["metric_type"].(string); ok && mt == "IP" {
		metricType = entity.IP // 使用内积距离可能对某些嵌入效果更好
	}

	// 创建高级搜索参数
	nprobe := 50 // 默认值更高
	if np, ok := params["params"].(map[string]interface{})["nprobe"].(int); ok {
		nprobe = np
	}

	sp, err := entity.NewIndexIvfFlatSearchParam(nprobe)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("创建搜索参数失败: %v", err)
	}

	// 创建搜索向量
	searchVector := entity.FloatVector(vector)
	searchVectors := []entity.Vector{searchVector}

	// 执行搜索
	results, err := MilvusClient.GetClient().Search(
		ctx,
		collName,
		[]string{},
		"",
		[]string{"content", "vector"},
		searchVectors,
		"vector",
		metricType,
		topK,
		sp,
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("搜索向量失败: %v", err)
	}

	// 处理结果与原函数相同
	if len(results) == 0 {
		return nil, nil, nil, fmt.Errorf("未找到匹配结果")
	}

	ids := make([]int64, 0, topK)
	scores := make([]float32, 0, topK)
	contents := make([]string, 0, topK)

	for _, result := range results {
		if result.IDs != nil {
			idCol, ok := result.IDs.(*entity.ColumnInt64)
			if !ok {
				return nil, nil, nil, fmt.Errorf("无效的ID类型")
			}
			ids = append(ids, idCol.Data()...)
			scores = append(scores, result.Scores...)
		}

		for _, field := range result.Fields {
			if contentCol, ok := field.(*entity.ColumnVarChar); ok {
				contents = append(contents, contentCol.Data()...)
			}
		}
	}

	return ids, scores, contents, nil
}

func DeleteMilvusCollection(ctx *context.Context) error {
	// 获取配置中的集合名称
	collectionName := config.GetConfig().Milvus.Collection

	// 初始化 Milvus 客户端（如果尚未初始化）
	if !IsClientInit() {
		if err := InitMilvus(ctx); err != nil {
			return fmt.Errorf("初始化 Milvus 客户端失败: %v", err)
		}
	}

	// 检查集合是否存在
	exists, err := MilvusClient.GetClient().HasCollection(*ctx, collectionName)
	if err != nil {
		return fmt.Errorf("检查集合是否存在失败: %v", err)
	}

	if !exists {
		return fmt.Errorf("集合 %s 不存在", collectionName)
	}

	// 删除集合
	if err := MilvusClient.GetClient().DropCollection(*ctx, collectionName); err != nil {
		return fmt.Errorf("删除集合 %s 失败: %v", collectionName, err)
	}

	log.Printf("成功删除集合: %s", collectionName)
	return nil
}

// LoadCollection 将集合加载到内存
func LoadCollection(ctx context.Context) error {
	collName := config.GetConfig().Milvus.Collection

	// 检查集合是否存在
	exist, err := MilvusClient.GetClient().HasCollection(ctx, collName)
	if err != nil {
		return fmt.Errorf("检查集合失败: %v", err)
	}

	if !exist {
		return fmt.Errorf("集合 %s 不存在", collName)
	}

	// 确保索引存在
	if err := CreateVectorIndex(ctx); err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}

	// 检查集合是否已加载
	loadStatus, err := MilvusClient.GetClient().GetLoadState(ctx, collName, []string{})
	if err != nil {
		return fmt.Errorf("获取集合加载状态失败: %v", err)
	}

	// 如果未加载，则加载集合
	if loadStatus != entity.LoadStateLoaded {
		err = MilvusClient.GetClient().LoadCollection(ctx, collName, false)
		if err != nil {
			return fmt.Errorf("加载集合失败: %v", err)
		}
		log.Printf("集合 %s 已成功加载到内存", collName)
	}

	return nil
}

// CreateVectorIndex 为向量字段创建索引
func CreateVectorIndex(ctx context.Context) error {
	collName := config.GetConfig().Milvus.Collection

	// 检查集合是否存在
	exist, err := MilvusClient.GetClient().HasCollection(ctx, collName)
	if err != nil {
		return fmt.Errorf("检查集合失败: %v", err)
	}

	if !exist {
		return fmt.Errorf("集合 %s 不存在", collName)
	}

	// 创建索引参数
	idx, err := entity.NewIndexIvfFlat(entity.L2, 1024) // nlist=1024
	if err != nil {
		return fmt.Errorf("创建索引参数失败: %v", err)
	}

	// 检查索引是否存在
	idxDesc, err := MilvusClient.GetClient().DescribeIndex(ctx, collName, "vector")
	if err == nil && len(idxDesc) > 0 {
		log.Printf("索引已存在于字段 'vector'")
		return nil
	}

	// 创建索引
	err = MilvusClient.GetClient().CreateIndex(ctx, collName, "vector", idx, false)
	if err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}

	log.Printf("成功为集合 %s 的 vector 字段创建索引", collName)
	return nil
}
