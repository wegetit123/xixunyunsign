package cmd_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

var role string

// 定义请求中使用的结构体
type ContentPart struct {
	Text string `json:"text"`
}

type Content struct {
	Parts []ContentPart `json:"parts"`
	Role  string        `json:"role"`
}

type Candidate struct {
	Content      Content `json:"content"`
	FinishReason string  `json:"finishReason"`
	AvgLogprobs  float64 `json:"avgLogprobs"`
}

type Response struct {
	Candidates    []Candidate `json:"candidates"`
	UsageMetadata interface{} `json:"usageMetadata"` // 根据需要定义具体类型
	ModelVersion  string      `json:"modelVersion"`
}

type RequestPayload struct {
	Contents []Content `json:"contents"`
}

func TestGenerateContent(t *testing.T) {
	// 从环境变量中获取 API 密钥
	apiKey := "AI*********"
	//if apiKey == "" {
	//	t.Fatal("未设置 GEMINI_API_KEY 环境变量")
	//}
	role = "********实习生"
	// 构建请求 URL
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=%s", apiKey)

	// 构建请求 payload
	payload := RequestPayload{
		Contents: []Content{
			{
				Role: "user",
				Parts: []ContentPart{
					{Text: fmt.Sprintf("我是%s。现在要求我回答作为第一个月的实习报告月报的回复。以替换content里的内容返回给我，以api的形式返回给我，不要回复其他的任何信息，不要```json和\\n", role)},
				},
			},
			{
				Role: "user",
				Parts: []ContentPart{
					{Text: "[{\"title\":\"实习工作具体情况及实习任务完成情况\",\"content\":\"\",\"require\":\"1\",\"sort\":1},{\"title\":\"主要收获及工作成绩\",\"content\":\"\",\"require\":\"0\",\"sort\":2},{\"title\":\"工作中的问题及需要老师的指导帮助\",\"content\":\"\",\"require\":\"0\",\"sort\":3}]"},
				},
			},
			//{
			//	Role: "user",
			//	Parts: []ContentPart{
			//		{Text: "Can you describe what the magic backpack looks like?"},
			//	},
			//},
		},
	}

	// 将 payload 序列化为 JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("JSON 序列化失败: %v", err)
	}

	// 创建 HTTP POST 请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("创建 HTTP 请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("发送 HTTP 请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("读取响应体失败: %v", err)
	}

	// 打印响应的所有内容
	t.Logf("响应状态码: %d", resp.StatusCode)
	t.Logf("响应体: %s", string(bodyBytes))

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("HTTP 请求失败，状态码: %s，响应体: %s", resp.Status, string(bodyBytes))
	}

	// 解析响应 JSON
	var responseData Response
	if err := json.Unmarshal(bodyBytes, &responseData); err != nil {
		t.Fatalf("解析响应 JSON 失败: %v", err)
	}

	// 检查是否有候选内容
	if len(responseData.Candidates) == 0 {
		t.Fatal("响应中不包含任何候选内容")
	}

	// 获取第一个候选的文本内容
	firstCandidate := responseData.Candidates[0]
	if len(firstCandidate.Content.Parts) == 0 {
		t.Fatal("第一个候选内容中不包含任何部分")
	}

	generatedText := firstCandidate.Content.Parts[0].Text

	// 打印生成的文本内容
	t.Logf("生成的文本内容: %s", generatedText)

	// 断言生成的文本不为空
	if generatedText == "" {
		t.Error("生成的文本内容为空")
	}

	//// 可根据需要添加更多断言，例如检查文本是否包含特定关键词
	//expectedSubstring := "AI works"
	//if !contains(generatedText, expectedSubstring) {
	//	t.Errorf("生成的文本不包含预期的子字符串: %s", expectedSubstring)
	//}
}

// contains 是一个简单的子字符串检查函数
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
