package sms

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v3/client"
	"github.com/alibabacloud-go/tea/tea"
)

// Client 短信服务客户端接口
type Client interface {
	// SendVerificationCode 发送验证码
	SendVerificationCode(ctx context.Context, phoneNumber string) (string, error)

	// VerifyCode 验证验证码
	VerifyCode(ctx context.Context, phoneNumber, code string) (bool, error)

	// Close 关闭客户端
	Close() error
}

// client 短信服务客户端实现
type client struct {
	config    *Config
	apiClient *dysmsapi.Client
}

// NewClient 创建短信服务客户端
func NewClient(config *Config) (Client, error) {
	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 创建阿里云 API 配置
	apiConfig := &openapi.Config{
		AccessKeyId:     tea.String(config.AccessKeyID),
		AccessKeySecret: tea.String(config.AccessKeySecret),
		RegionId:        tea.String(config.Region),
		Endpoint:        tea.String(fmt.Sprintf("dysmsapi.aliyuncs.com")),
	}

	// 创建短信 API 客户端
	apiClient, err := dysmsapi.NewClient(apiConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create sms api client: %w", err)
	}

	return &client{
		config:    config,
		apiClient: apiClient,
	}, nil
}

// SendVerificationCode 发送验证码
func (c *client) SendVerificationCode(ctx context.Context, phoneNumber string) (string, error) {
	// 生成 6 位随机验证码
	code := generateVerificationCode(CodeLength)

	// 构建发送请求
	request := &dysmsapi.SendSmsRequest{
		PhoneNumbers:  tea.String(phoneNumber),
		SignName:      tea.String(c.config.SignName),
		TemplateCode:  tea.String(c.config.TemplateCode),
		TemplateParam: tea.String(fmt.Sprintf(`{"code":"%s"}`, code)),
	}

	// 发送短信
	response, err := c.apiClient.SendSms(request)
	if err != nil {
		return "", fmt.Errorf("failed to send sms: %w", err)
	}

	// 检查响应
	if response.Body == nil {
		return "", fmt.Errorf("empty response body")
	}

	if response.Body.Code == nil || *response.Body.Code != "OK" {
		errMsg := "unknown error"
		if response.Body.Message != nil {
			errMsg = *response.Body.Message
		}
		return "", fmt.Errorf("sms send failed: %s", errMsg)
	}

	return code, nil
}

// VerifyCode 验证验证码
func (c *client) VerifyCode(ctx context.Context, phoneNumber, code string) (bool, error) {
	// 注意：实际的验证码验证逻辑应该在 Service 层通过 Redis 实现
	// 这里只是一个占位方法，实际使用时应该调用 Service 层的验证方法
	return false, fmt.Errorf("verification should be done in service layer with redis")
}

// Close 关闭客户端
func (c *client) Close() error {
	// 阿里云 SDK 客户端不需要显式关闭
	return nil
}

// generateVerificationCode 生成指定长度的数字验证码
func generateVerificationCode(length int) string {
	rand.Seed(time.Now().UnixNano())
	code := ""
	for i := 0; i < length; i++ {
		code += fmt.Sprintf("%d", rand.Intn(10))
	}
	return code
}
