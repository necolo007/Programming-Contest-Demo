package bocha

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	BaseURL = "https://api.bochaai.com/v1"
)

// Client 博查AI客户端
type Client struct {
	APIKey     string
	HTTPClient *http.Client
}

// 创建新的博查AI客户端
func NewClient(apiKey string) *Client {
	return &Client{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// SearchRequest 搜索请求结构
type SearchRequest struct {
	Query     string `json:"query"`     // 用户的搜索词
	Freshness string `json:"freshness"` // 时间范围: oneDay, oneWeek, oneMonth, oneYear, noLimit(默认)
	Summary   bool   `json:"summary"`   // 是否显示文本摘要
	Count     int    `json:"count"`     // 返回结果条数，范围1-50，默认10
	Page      int    `json:"page"`      // 页码，默认1
}

// 响应相关结构体定义
type QueryContext struct {
	OriginalQuery string `json:"originalQuery"`
}

type WebPageValue struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	URL              string `json:"url"`
	DisplayURL       string `json:"displayUrl"`
	Snippet          string `json:"snippet"`
	Summary          string `json:"summary,omitempty"`
	SiteName         string `json:"siteName"`
	SiteIcon         string `json:"siteIcon"`
	DateLastCrawled  string `json:"dateLastCrawled"`
	CachedPageURL    string `json:"cachedPageUrl"`
	Language         string `json:"language"`
	IsFamilyFriendly bool   `json:"isFamilyFriendly"`
	IsNavigational   bool   `json:"isNavigational"`
}

type WebSearchWebPages struct {
	WebSearchURL          string         `json:"webSearchUrl"`
	TotalEstimatedMatches int            `json:"totalEstimatedMatches"`
	Value                 []WebPageValue `json:"value"`
	SomeResultsRemoved    bool           `json:"someResultsRemoved"`
}

type Thumbnail struct {
	Height int `json:"height"`
	Width  int `json:"width"`
}

type ImageValue struct {
	WebSearchURL       string    `json:"webSearchUrl"`
	Name               string    `json:"name"`
	ThumbnailURL       string    `json:"thumbnailUrl"`
	DatePublished      string    `json:"datePublished"`
	ContentURL         string    `json:"contentUrl"`
	HostPageURL        string    `json:"hostPageUrl"`
	ContentSize        string    `json:"contentSize"`
	EncodingFormat     string    `json:"encodingFormat"`
	HostPageDisplayURL string    `json:"hostPageDisplayUrl"`
	Width              int       `json:"width"`
	Height             int       `json:"height"`
	Thumbnail          Thumbnail `json:"thumbnail"`
}

type WebSearchImages struct {
	ID               string       `json:"id"`
	ReadLink         string       `json:"readLink"`
	WebSearchURL     string       `json:"webSearchUrl"`
	IsFamilyFriendly bool         `json:"isFamilyFriendly"`
	Value            []ImageValue `json:"value"`
}

type Publisher struct {
	Name string `json:"name"`
}

type Creator struct {
	Name string `json:"name"`
}

type VideoValue struct {
	WebSearchURL       string      `json:"webSearchUrl"`
	Name               string      `json:"name"`
	Description        string      `json:"description"`
	ThumbnailURL       string      `json:"thumbnailUrl"`
	Publisher          []Publisher `json:"publisher"`
	Creator            Creator     `json:"creator"`
	ContentURL         string      `json:"contentUrl"`
	HostPageURL        string      `json:"hostPageUrl"`
	EncodingFormat     string      `json:"encodingFormat"`
	HostPageDisplayURL string      `json:"hostPageDisplayUrl"`
	Width              int         `json:"width"`
	Height             int         `json:"height"`
	Duration           string      `json:"duration"`
	MotionThumbnailURL string      `json:"motionThumbnailUrl"`
	EmbedHTML          string      `json:"embedHtml"`
	AllowHTTPSEmbed    bool        `json:"allowHttpsEmbed"`
	ViewCount          int         `json:"viewCount"`
	Thumbnail          Thumbnail   `json:"thumbnail"`
	AllowMobileEmbed   bool        `json:"allowMobileEmbed"`
	IsSuperFresh       bool        `json:"isSuperfresh"`
	DatePublished      string      `json:"datePublished"`
}

type WebSearchVideos struct {
	ID               string       `json:"id"`
	ReadLink         string       `json:"readLink"`
	WebSearchURL     string       `json:"webSearchUrl"`
	IsFamilyFriendly bool         `json:"isFamilyFriendly"`
	Scenario         string       `json:"scenario"`
	Value            []VideoValue `json:"value"`
}

type SearchData struct {
	Type         string            `json:"_type"`
	QueryContext QueryContext      `json:"queryContext"`
	WebPages     WebSearchWebPages `json:"webPages"`
	Images       WebSearchImages   `json:"images"`
	Videos       WebSearchVideos   `json:"videos"`
}

type SearchResponse struct {
	Code  int        `json:"code"`
	LogID string     `json:"log_id"`
	Msg   string     `json:"msg"`
	Data  SearchData `json:"data"`
}

// Search 执行博查搜索
func (c *Client) Search(req SearchRequest) (string, error) {
	// 设置默认值
	if req.Freshness == "" {
		req.Freshness = "noLimit"
	}
	if req.Count <= 0 || req.Count > 50 {
		req.Count = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	// 将请求转换为JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", BaseURL+"/web-search", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	// 执行请求
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("执行HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应内容失败: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API错误: %s, 状态码: %d", string(body), resp.StatusCode)
	}

	return string(body), nil
}

// ExtractSearchInfo 从搜索结果JSON中提取有用信息
func ExtractSearchInfo(searchResultJSON string) (string, error) {
	var searchResp SearchResponse
	if err := json.Unmarshal([]byte(searchResultJSON), &searchResp); err != nil {
		return "", fmt.Errorf("解析搜索结果失败: %w", err)
	}

	if searchResp.Code != 200 {
		return "", fmt.Errorf("搜索API返回错误: %s, 代码: %d", searchResp.Msg, searchResp.Code)
	}

	var builder strings.Builder

	// 添加搜索原始查询
	builder.WriteString("搜索查询: ")
	builder.WriteString(searchResp.Data.QueryContext.OriginalQuery)
	builder.WriteString("\n\n")

	// 如果没有搜索结果
	if len(searchResp.Data.WebPages.Value) == 0 {
		builder.WriteString("未找到相关网页。")
		return builder.String(), nil
	}

	// 添加搜索结果总数信息
	builder.WriteString(fmt.Sprintf("找到约 %d 个相关结果\n\n", searchResp.Data.WebPages.TotalEstimatedMatches))

	// 添加网页结果
	for i, page := range searchResp.Data.WebPages.Value {
		builder.WriteString(fmt.Sprintf("--- 结果 %d ---\n", i+1))
		builder.WriteString(fmt.Sprintf("标题: %s\n", page.Name))
		builder.WriteString(fmt.Sprintf("网址: %s\n", page.URL))
		builder.WriteString(fmt.Sprintf("来源: %s\n", page.SiteName))

		// 添加日期（如果有）
		if page.DateLastCrawled != "" {
			// 处理时间格式 2025-02-23T08:18:30Z
			dateStr := page.DateLastCrawled
			dateStr = strings.TrimSuffix(dateStr, "Z")
			if len(dateStr) > 10 {
				dateStr = dateStr[:10] // 只保留日期部分
			}
			builder.WriteString(fmt.Sprintf("日期: %s\n", dateStr))
		}

		// 添加摘要
		if page.Summary != "" {
			builder.WriteString(fmt.Sprintf("摘要: %s\n", page.Summary))
		} else if page.Snippet != "" {
			builder.WriteString(fmt.Sprintf("摘要: %s\n", page.Snippet))
		}

		builder.WriteString("\n")

		// 限制结果数量，避免内容过长
		if i >= 4 {
			builder.WriteString(fmt.Sprintf("还有 %d 个结果未显示。\n", len(searchResp.Data.WebPages.Value)-5))
			break
		}
	}

	return builder.String(), nil
}
