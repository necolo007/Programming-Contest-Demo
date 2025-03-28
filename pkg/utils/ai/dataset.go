package ai

import (
	"Programming-Demo/core/milvus"
	"encoding/csv"
	"fmt"
	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"
)

// ImportCSVToMilvus 优化后的CSV导入函数
func ImportCSVToMilvus(filepath string, batchSize int) error {
	// 1. 打开并读取CSV文件
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("无法打开CSV文件: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("读取CSV内容失败: %v", err)
	}

	// 跳过标题行
	records = records[1:]
	total := len(records)
	color.Blue("共读取 %d 条记录", total)

	// 2. 创建进度条
	bar := progressbar.NewOptions(
		total,
		progressbar.OptionSetDescription("导入进度"),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
	)

	// 3. 小批量处理，每批10条记录
	successCount := 0

	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}

		batch := records[i:end]

		// 4. 处理当前批次
		if err := processBatchWithRetry(batch, int64(i)); err != nil {
			color.Red("处理批次 %d 失败: %v", i/batchSize+1, err)
			// 继续处理其他批次
		} else {
			successCount += len(batch)
		}

		// 5. 更新进度条
		bar.Add(len(batch))

		// 6. 添加延迟防止API限流
		time.Sleep(3 * time.Second)
	}

	color.Green("成功导入 %d/%d 条记录到 Milvus", successCount, total)
	return nil
}

// processBatchWithRetry 使用指数退避重试处理批次
func processBatchWithRetry(records [][]string, startID int64) error {
	maxRetries := 5

	for attempt := 0; attempt < maxRetries; attempt++ {
		// 为重试添加指数退避延迟
		if attempt > 0 {
			// 退避时间 = 2^尝试次数 * (800 + [0, 400)ms)重试抖动
			backoff := time.Duration(math.Pow(2, float64(attempt))*float64(800+rand.Intn(400))) * time.Millisecond
			color.Red("重试 %d/%d，等待 %v...", attempt+1, maxRetries, backoff)
			time.Sleep(backoff)
		}

		err := processBatch(records, startID)
		if err == nil {
			return nil
		}

		if strings.Contains(err.Error(), "Throttling.User") {
			// 对于用户限流错误，使用更长的退避时间
			backoff := time.Duration(math.Pow(2, float64(attempt+2))) * time.Second
			color.Red("用户请求限流，等待较长时间: %v", backoff)
			time.Sleep(backoff)
			continue
		}

		// 处理速率限制相关的错误
		if strings.Contains(err.Error(), "ServiceUnavailable") ||
			strings.Contains(err.Error(), "Throttling") ||
			strings.Contains(err.Error(), "RequestThrottled") ||
			strings.Contains(err.Error(), "TooManyRequests") {
			color.Yellow("API请求被限制，正在重试: %v", err)
			continue
		}

		// 对于其他错误，减少重试次数
		if attempt >= 2 {
			return err
		}
	}

	return fmt.Errorf("达到最大重试次数 %d 后仍失败", maxRetries)
}

// processBatch 优化后的批处理函数
func processBatch(records [][]string, startID int64) error {
	if len(records) == 0 {
		return nil
	}

	// 准备数据
	contents := make([]string, 0, len(records))
	ids := make([]int64, 0, len(records))
	validRecords := 0

	// 首先计算有效记录数
	for i, record := range records {
		if len(record) >= 2 {
			contents = append(contents, fmt.Sprintf("法条编章：%s法条内容：%s", record[0]+record[1]+record[2], record[3]+record[4]))
			ids = append(ids, startID+int64(i))
			validRecords++
		}
	}

	// 预分配向量数组，确保维度一致
	vectors := make([][]float32, validRecords)

	// 单独处理向量生成
	validIdx := 0
	for _, record := range records {
		if len(record) < 2 {
			continue
		}

		// 只为问题生成向量
		embedding, err := GenerateEmbedding(record[0])
		if err != nil {
			color.Red("向量生成失败，跳过此条：%v", err)
			// 移除对应的content和id
			if validIdx < len(contents) {
				contents = append(contents[:validIdx], contents[validIdx+1:]...)
				ids = append(ids[:validIdx], ids[validIdx+1:]...)
			}
			continue
		}

		// 转换为float32
		embedding32 := make([]float32, len(embedding))
		for j, val := range embedding {
			embedding32[j] = float32(val)
		}

		vectors[validIdx] = embedding32
		validIdx++
	}

	// 调整最终数组大小
	vectors = vectors[:validIdx]

	// 确保向量和内容数量一致
	if len(vectors) != len(contents) {
		return fmt.Errorf("向量数量(%d)与内容数量(%d)不匹配", len(vectors), len(contents))
	}

	// 打印向量维度信息以便调试
	if len(vectors) > 0 {
		color.Blue("向量维度: %d条记录，每条向量%d维", len(vectors), len(vectors[0]))
	}

	// 插入到Milvus
	if err := milvus.InsertVectors(vectors, contents); err != nil {
		return fmt.Errorf("插入Milvus失败: %v", err)
	}

	color.Green("成功插入%d条记录", len(contents))
	return nil
}
