package ai

import (
	"Programming-Demo/core/milvus"
	"encoding/csv" // 在 main 函数或应用启动处添加
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// ImportCSVToMilvus 将CSV文件数据导入到Milvus
func ImportCSVToMilvus(filePath string, batchSize int) error {
	// 打开CSV文件
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("无法打开文件: %v", err)
	}
	defer file.Close()

	// 配置CSV读取器
	reader := csv.NewReader(file)
	reader.Comma = ','
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	// 读取所有记录
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("读取CSV文件失败: %v", err)
	}

	log.Printf("读取到 %d 条记录，开始处理...", len(records))

	// 批量处理
	totalCount := len(records)
	processedCount := 0
	startTime := time.Now()

	for i := 0; i < totalCount; i += batchSize {
		endIndex := i + batchSize
		if endIndex > totalCount {
			endIndex = totalCount
		}

		batch := records[i:endIndex]
		err = processBatch(batch, int64(i+1))
		if err != nil {
			return fmt.Errorf("处理批次 %d-%d 失败: %v", i, endIndex-1, err)
		}

		processedCount += len(batch)
		elapsed := time.Since(startTime)
		log.Printf("已处理 %d/%d 条记录 (%.2f%%), 耗时: %v",
			processedCount, totalCount, float64(processedCount)/float64(totalCount)*100, elapsed)
	}

	log.Printf("数据导入完成，总耗时: %v", time.Since(startTime))
	return nil
}

// processBatch 处理一批记录
func processBatch(records [][]string, startID int64) error {
	var vectors [][]float32
	var ids []int64
	var contents []string

	for i, record := range records {
		if len(record) < 2 {
			continue
		}

		// 清洗和处理文本
		question := cleanText(record[0])
		answer := ""
		if len(record) > 1 {
			answer = cleanText(record[1])
		}

		// 合并问题和答案作为内容
		content := fmt.Sprintf("问题: %s\n答案: %s", question, answer)

		// 生成向量嵌入
		embedding, err := GenerateEmbedding(question)
		if err != nil {
			log.Printf("生成向量嵌入失败: %v，跳过该记录", err)
			continue
		}

		// 转换向量类型 float64 -> float32
		vector := make([]float32, len(embedding))
		for j, v := range embedding {
			vector[j] = float32(v)
		}

		vectors = append(vectors, vector)
		contents = append(contents, content)
		ids = append(ids, startID+int64(i))
	}

	// 检查是否有数据需要插入
	if len(vectors) == 0 {
		log.Printf("没有可用的向量数据需要插入")
		return nil
	}

	log.Printf("开始插入%d向量数据...", len(vectors))
	if !milvus.IsClientInit() {
		return fmt.Errorf("milvus客户端未正确初始化")
	}
	// 插入向量数据到Milvus
	// 注意: InsertVectors参数顺序为 vectors, contents
	err := milvus.InsertVectors(vectors, contents)
	if err != nil {
		return fmt.Errorf("插入向量数据失败: %v", err)
	}

	return nil
}

// cleanText 清洗文本
func cleanText(text string) string {
	// 去除多余空白字符
	text = strings.TrimSpace(text)
	// 去除重复的换行符
	text = strings.ReplaceAll(text, "\n\n", "\n")
	return text
}
