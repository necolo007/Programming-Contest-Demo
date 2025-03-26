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
	"sync"
	"time"
)

// ImportCSVToMilvusWithThrottle 导入CSV数据到Milvus，同时对阿里云API实施适当速率限制
func ImportCSVToMilvusWithThrottle(filepath string, batchSize int) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("无法打开CSV文件: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ','
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("读取CSV内容失败: %v", err)
	}

	// 跳过标题行
	records = records[1:]
	total := len(records)
	color.Blue("共读取 %d 条记录", total)

	// 创建速率限制器 - 优化为15 QPS
	rateLimiter := time.NewTicker(250 * time.Millisecond)
	defer rateLimiter.Stop()

	// 并发控制 - 根据阿里云API实际限制调整
	sem := make(chan struct{}, 6)

	// 错误和完成通道
	errChan := make(chan error, 1)
	doneChan := make(chan bool, 1)

	totalBatches := (total + batchSize - 1) / batchSize
	processedBatches := 0
	var mu sync.Mutex // 用于保护processedBatches

	// 创建进度条
	// 创建进度条
	bar := progressbar.NewOptions(total,
		progressbar.OptionSetDescription("导入进度"),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
	)

	// 分批处理
	for i := 0; i < total; i += batchSize {
		batchEnd := i + batchSize
		if batchEnd > total {
			batchEnd = total
		}

		batch := records[i:batchEnd]
		startID := int64(i)
		batchNum := (i / batchSize) + 1

		// 等待速率限制器的信号
		<-rateLimiter.C

		// 获取信号量槽位
		sem <- struct{}{}

		go func(b [][]string, id int64, num int, size int) {
			defer func() { <-sem }() // 完成时释放信号量

			if err := processBatchWithRetry(b, id); err != nil {
				select {
				case errChan <- fmt.Errorf("处理批次 %d 失败: %v", num, err):
				default:
					// 避免阻塞
				}
				return
			}

			// 更新进度
			bar.Add(size)

			mu.Lock()
			processedBatches++
			currentProgress := processedBatches
			mu.Unlock()

			if currentProgress%10 == 0 || currentProgress == totalBatches {
				color.Green("已完成 %d/%d 批次 (%.1f%%)",
					currentProgress, totalBatches, float64(currentProgress)/float64(totalBatches)*100)
			}

			if currentProgress == totalBatches {
				doneChan <- true
			}
		}(batch, startID, batchNum, len(batch))
	}

	// 等待完成或错误
	select {
	case err := <-errChan:
		return err
	case <-doneChan:
		color.Green("成功导入 %d 条记录到 Milvus", total)
		return nil
	}
}

// processBatchWithRetry 使用指数退避重试处理批次
func processBatchWithRetry(records [][]string, startID int64) error {
	maxRetries := 5

	for attempt := 0; attempt < maxRetries; attempt++ {
		// 为重试添加指数退避延迟
		if attempt > 0 {
			// 退避时间 = 2^尝试次数 * (800 + [0, 400)ms)重试抖动
			backoff := time.Duration(math.Pow(2, float64(attempt))*float64(800+rand.Intn(400))) * time.Millisecond
			color.Yellow("重试 %d/%d，等待 %v...", attempt+1, maxRetries, backoff)
			time.Sleep(backoff)
		}

		err := processBatch(records, startID)
		if err == nil {
			return nil
		}

		if strings.Contains(err.Error(), "Throttling.User") {
			// 对于用户限流错误，使用更长的退避时间
			backoff := time.Duration(math.Pow(2, float64(attempt+2))) * time.Second
			color.Yellow("用户请求限流，等待较长时间: %v", backoff)
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

// processBatch 使用子批次处理单个批次
func processBatch(records [][]string, startID int64) error {
	// 准备嵌入数据
	questions := make([]string, 0, len(records))
	answers := make([]string, 0, len(records))
	ids := make([]int64, 0, len(records))

	for i, record := range records {
		if len(record) < 2 {
			continue
		}
		questions = append(questions, record[0])
		answers = append(answers, record[1])
		ids = append(ids, startID+int64(i))
	}

	// 使用更小的子批次（每次API调用2项）
	subBatchSize := 2
	vectors := make([][]float32, 0, len(questions))

	// 顺序处理每个子批次，中间添加短暂延迟
	for i := 0; i < len(questions); i += subBatchSize {
		end := i + subBatchSize
		if end > len(questions) {
			end = len(questions)
		}

		// 提取当前子批次的问题
		currentBatch := questions[i:end]

		// 为每个问题单独生成嵌入向量
		for _, question := range currentBatch {
			// GenerateEmbedding 接受单个字符串并返回 []float64
			embedding, err := GenerateEmbedding(question)
			if err != nil {
				return fmt.Errorf("获取嵌入向量失败: %v", err)
			}

			// 将 []float64 转换为 []float32
			embedding32 := make([]float32, len(embedding))
			for j, val := range embedding {
				embedding32[j] = float32(val)
			}

			// 正确地将向量作为一个元素追加到vectors中
			vectors = append(vectors, embedding32)
		}

		// 在子批次之间添加小延迟
		if end < len(questions) {
			time.Sleep(500 * time.Millisecond)
		}
	}

	// 插入到Milvus
	textData := make([]string, len(questions)+len(answers))
	copy(textData, questions)
	copy(textData[len(questions):], answers)

	if err := milvus.InsertVectors(vectors, textData); err != nil {
		return fmt.Errorf("插入Milvus失败: %v", err)
	}

	return nil
}
