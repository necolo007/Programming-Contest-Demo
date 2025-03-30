package ai_service

import (
	"Programming-Demo/core/gin/dbs"
	"Programming-Demo/internal/app/ai/ai_entity"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/ristretto"
	"time"
)

var (
	// 全局Ristretto缓存实例
	cacheInstance *ristretto.Cache
)

const (
	DefaultHistoryLimit = 50             // 默认历史记录限制数量
	CacheTTL            = 24 * time.Hour // 缓存生存时间
	CacheMaxCost        = int64(1 << 30) // 最大缓存大小（1GB）
	CacheNumCounters    = int64(1e7)     // 估计存储键数量
)

// 初始化缓存
func InitCache() error {
	if cacheInstance != nil {
		return nil // 缓存已初始化
	}

	var err error
	// 配置Ristretto缓存
	config := &ristretto.Config{
		NumCounters: CacheNumCounters, // 大约占用 10MB 空间
		MaxCost:     CacheMaxCost,     // 最大 1GB
		BufferItems: 64,               // 提高性能的缓冲区大小
		Metrics:     true,             // 启用指标收集
	}

	cacheInstance, err = ristretto.NewCache(config)
	if err != nil {
		return fmt.Errorf("failed to initialize ristretto cache: %w", err)
	}

	return nil
}

// 生成缓存键
func getCacheKey(userID uint, theme string) string {
	return fmt.Sprintf("chat:%d:%s", userID, theme)
}

// 保存聊天缓存到Ristretto
func SaveChatCache(userID uint, theme string, messages []ai_entity.ChatHistory) error {
	if err := InitCache(); err != nil {
		return err
	}

	// 创建缓存对象
	cache := ai_entity.LocalChatCache{
		UserID:      userID,
		Theme:       theme,
		LastUpdated: time.Now(),
		Messages:    messages,
		Metadata: map[string]interface{}{
			"count": len(messages),
		},
	}

	// 将结构体序列化为JSON字节，以便于缓存存储
	data, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	// 存储到Ristretto缓存
	cacheKey := getCacheKey(userID, theme)
	success := cacheInstance.SetWithTTL(cacheKey, data, 1, CacheTTL)
	if !success {
		return fmt.Errorf("failed to save data to cache")
	}

	// 确保写入完成
	cacheInstance.Wait()
	return nil
}

// 从Ristretto加载聊天缓存
func LoadChatCache(userID uint, theme string) (*ai_entity.LocalChatCache, error) {
	if err := InitCache(); err != nil {
		return nil, err
	}

	cacheKey := getCacheKey(userID, theme)
	cachedValue, found := cacheInstance.Get(cacheKey)
	if !found {
		return nil, nil // 缓存未命中
	}

	// 将缓存的JSON数据转换回结构体
	data, ok := cachedValue.([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid cache format")
	}

	var cache ai_entity.LocalChatCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return &cache, nil
}

// 删除聊天缓存
func DeleteChatCache(userID uint, theme string) error {
	if err := InitCache(); err != nil {
		return err
	}

	cacheKey := getCacheKey(userID, theme)
	cacheInstance.Del(cacheKey)
	return nil
}

// 清除所有缓存（通常在服务重启或维护时使用）
func ClearAllCache() {
	if cacheInstance != nil {
		cacheInstance.Clear()
	}
}

// 检查缓存是否有效（未过期）
func IsCacheValid(cache *ai_entity.LocalChatCache, maxAge time.Duration) bool {
	if cache == nil {
		return false
	}
	return time.Since(cache.LastUpdated) <= maxAge
}

// 获取指定主题的最近聊天记录
func GetRecentChatHistoryByTheme(userID uint, theme string, limit int) ([]ai_entity.ChatHistory, error) {
	if limit <= 0 {
		limit = DefaultHistoryLimit
	}

	var histories []ai_entity.ChatHistory
	if err := dbs.DB.Where("user_id = ? AND theme = ?", userID, theme).
		Order("created_at DESC").
		Limit(limit).
		Find(&histories).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent chat history: %w", err)
	}

	// 反转顺序以按时间正序排列
	for i, j := 0, len(histories)-1; i < j; i, j = i+1, j-1 {
		histories[i], histories[j] = histories[j], histories[i]
	}

	return histories, nil
}

// 更新现有主题的最后消息时间或创建新主题
func UpdateOrCreateTheme(userID uint, themeName string) error {
	now := time.Now()

	// 尝试查找现有主题
	var count int64
	dbs.DB.Model(&ai_entity.ChatTheme{}).
		Where("user_id = ? AND theme = ?", userID, themeName).
		Count(&count)

	if count == 0 {
		// 创建新主题
		theme := ai_entity.ChatTheme{
			UserID:      userID,
			Theme:       themeName,
			LastMessage: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		return dbs.DB.Create(&theme).Error
	}

	// 更新现有主题
	return dbs.DB.Exec(`
		UPDATE chat_themes 
		SET last_message = ?, updated_at = ? 
		WHERE user_id = ? AND theme = ?
	`, now, now, userID, themeName).Error
}

// 按时间正序获取指定主题的聊天历史
func GetChatHistoryByTheme(userID uint, theme string, limit int) ([]ai_entity.ChatHistory, error) {
	if limit <= 0 {
		limit = DefaultHistoryLimit
	}

	var histories []ai_entity.ChatHistory
	if err := dbs.DB.Where("user_id = ? AND theme = ?", userID, theme).
		Order("created_at ASC").
		Limit(limit).
		Find(&histories).Error; err != nil {
		return nil, fmt.Errorf("failed to get chat history: %w", err)
	}

	return histories, nil
}

// 获取用户的所有聊天主题
func GetAllChatThemes(userID uint) ([]ai_entity.ChatTheme, error) {
	var themes []ai_entity.ChatTheme
	if err := dbs.DB.Where("user_id = ?", userID).
		Order("last_message DESC").
		Find(&themes).Error; err != nil {
		return nil, fmt.Errorf("failed to get chat themes: %w", err)
	}

	return themes, nil
}
