package video

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SeedanceClient Seedance 2.0 视频生成客户端
type SeedanceClient struct {
	BaseURL       string
	APIKey        string
	Model         string
	Endpoint      string
	QueryEndpoint string
	HTTPClient    *http.Client
}

// SeedanceContent 视频生成内容
type SeedanceContent struct {
	Type     string                 `json:"type"`
	Text     string                 `json:"text,omitempty"`
	ImageURL map[string]interface{} `json:"image_url,omitempty"`
	Role     string                 `json:"role,omitempty"`
}

// SeedanceRequest 视频生成请求
type SeedanceRequest struct {
	Model           string             `json:"model"`
	Content         []SeedanceContent  `json:"content"`
	Duration        int                `json:"duration,omitempty"`
	AspectRatio     string             `json:"aspect_ratio,omitempty"`
	Seed            int64              `json:"seed,omitempty"`
	NegativePrompt  string             `json:"negative_prompt,omitempty"`
}

// SeedanceResponse 视频生成响应
type SeedanceResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Status  string `json:"status"`
	Error   struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
	CreatedAt int64 `json:"created"`
	Input     struct {
		Prompt string `json:"prompt"`
	} `json:"input,omitempty"`
	Output struct {
		VideoURL string `json:"video_url,omitempty"`
	} `json:"output,omitempty"`
}

func NewSeedanceClient(baseURL, apiKey, model, endpoint, queryEndpoint string) *SeedanceClient {
	if endpoint == "" {
		endpoint = "/video/generation"
	}
	if queryEndpoint == "" {
		queryEndpoint = "/video/generation/{taskId}"
	}
	return &SeedanceClient{
		BaseURL:       baseURL,
		APIKey:        apiKey,
		Model:         model,
		Endpoint:      endpoint,
		QueryEndpoint: queryEndpoint,
		HTTPClient: &http.Client{
			Timeout: 300 * time.Second,
		},
	}
}

// GenerateVideo 生成视频（支持文生视频和图生视频）
func (c *SeedanceClient) GenerateVideo(imageURL, prompt string, opts ...VideoOption) (*VideoResult, error) {
	options := &VideoOptions{
		Duration:    5,
		AspectRatio: "16:9",
	}

	for _, opt := range opts {
		opt(options)
	}

	model := c.Model
	if options.Model != "" {
		model = options.Model
	}

	// 构建prompt文本
	promptText := prompt
	if options.AspectRatio != "" {
		promptText += fmt.Sprintf("  --ratio %s", options.AspectRatio)
	}
	if options.Duration > 0 {
		promptText += fmt.Sprintf("  --dur %d", options.Duration)
	}

	content := []SeedanceContent{
		{
			Type: "text",
			Text: promptText,
		},
	}

	// 处理图生视频模式
	if imageURL != "" {
		content = append(content, SeedanceContent{
			Type: "image_url",
			ImageURL: map[string]interface{}{
				"url": imageURL,
			},
		})
	}

	// 构建请求体
	reqBody := SeedanceRequest{
		Model:   model,
		Content: content,
	}

	// 添加可选参数
	if options.Duration > 0 {
		reqBody.Duration = options.Duration
	}
	if options.AspectRatio != "" {
		reqBody.AspectRatio = options.AspectRatio
	}
	if options.Seed > 0 {
		reqBody.Seed = options.Seed
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	endpoint := c.BaseURL + c.Endpoint
	fmt.Printf("[Seedance] Generating video - Endpoint: %s, FullURL: %s, Model: %s\n", c.Endpoint, endpoint, model)
	fmt.Printf("[Seedance] Request body: %s\n", string(jsonData))

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	fmt.Printf("[Seedance] Response status: %d, body: %s\n", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result SeedanceResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	fmt.Printf("[Seedance] Video generation initiated - TaskID: %s, Status: %s\n", result.ID, result.Status)

	if result.Error.Message != "" {
		return nil, fmt.Errorf("seedance error: %s", result.Error.Message)
	}

	videoResult := &VideoResult{
		TaskID:    result.ID,
		Status:    result.Status,
		Completed: result.Status == "completed" || result.Status == "succeeded",
	}

	if result.Output.VideoURL != "" {
		videoResult.VideoURL = result.Output.VideoURL
		videoResult.Completed = true
	}

	return videoResult, nil
}

// GetTaskStatus 查询任务状态
func (c *SeedanceClient) GetTaskStatus(taskID string) (*VideoResult, error) {
	// 替换占位符
	queryPath := c.QueryEndpoint
	if strings.Contains(queryPath, "{taskId}") {
		queryPath = strings.ReplaceAll(queryPath, "{taskId}", taskID)
	} else if strings.Contains(queryPath, "{task_id}") {
		queryPath = strings.ReplaceAll(queryPath, "{task_id}", taskID)
	} else {
		queryPath = queryPath + "/" + taskID
	}

	endpoint := c.BaseURL + queryPath
	fmt.Printf("[Seedance] Querying task status - TaskID: %s, QueryEndpoint: %s, FullURL: %s\n", taskID, c.QueryEndpoint, endpoint)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	fmt.Printf("[Seedance] Response body: %s\n", string(body))

	var result SeedanceResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	fmt.Printf("[Seedance] Parsed result - ID: %s, Status: %s, VideoURL: %s\n", result.ID, result.Status, result.Output.VideoURL)

	videoResult := &VideoResult{
		TaskID:    result.ID,
		Status:    result.Status,
		Completed: result.Status == "completed" || result.Status == "succeeded",
	}

	if result.Error.Message != "" {
		videoResult.Error = result.Error.Message
	}

	if result.Output.VideoURL != "" {
		videoResult.VideoURL = result.Output.VideoURL
		videoResult.Completed = true
	}

	return videoResult, nil
}
