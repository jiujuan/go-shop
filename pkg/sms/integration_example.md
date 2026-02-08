# SMS 服务集成示例

本文档展示如何在 Go-Shop 应用中集成和使用短信服务。

## 1. 在主程序中初始化 SMS 客户端

在 `cmd/server/main.go` 中初始化 SMS 客户端：

```go
package main

import (
    "log"
    
    "go-shop/config"
    "go-shop/pkg/sms"
)

func main() {
    // 加载配置
    cfg, err := config.LoadConfig("config/config.yaml")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // 初始化 SMS 客户端（如果配置了）
    if cfg.SMS.AccessKeyID != "" {
        smsConfig := &sms.Config{
            AccessKeyID:     cfg.SMS.AccessKeyID,
            AccessKeySecret: cfg.SMS.AccessKeySecret,
            SignName:        cfg.SMS.SignName,
            TemplateCode:    cfg.SMS.TemplateCode,
            Region:          cfg.SMS.Region,
        }
        
        if err := sms.Init(smsConfig); err != nil {
            log.Printf("Warning: Failed to init SMS client: %v", err)
        } else {
            log.Println("SMS client initialized successfully")
        }
    }
    
    // ... 其他初始化代码
}
```

## 2. 创建 SMS Service 层

创建 `internal/service/sms_service.go`：

```go
package service

import (
    "context"
    "fmt"
    "time"
    
    "go-shop/pkg/sms"
    
    "github.com/redis/go-redis/v9"
)

// SMSService 短信服务接口
type SMSService interface {
    // SendVerificationCode 发送验证码
    SendVerificationCode(ctx context.Context, phone string) error
    
    // VerifyCode 验证验证码
    VerifyCode(ctx context.Context, phone, code string) (bool, error)
    
    // CheckRateLimit 检查发送频率限制
    CheckRateLimit(ctx context.Context, phone string) (bool, error)
}

// smsService 短信服务实现
type smsService struct {
    smsClient sms.Client
    redis     *redis.Client
}

// NewSMSService 创建短信服务实例
func NewSMSService(redis *redis.Client) (SMSService, error) {
    client, err := sms.GetClient()
    if err != nil {
        return nil, fmt.Errorf("failed to get sms client: %w", err)
    }
    
    return &smsService{
        smsClient: client,
        redis:     redis,
    }, nil
}

// SendVerificationCode 发送验证码
func (s *smsService) SendVerificationCode(ctx context.Context, phone string) error {
    // 1. 检查发送频率限制（60 秒内不能重复发送）
    sentKey := sms.RedisKeySentPrefix + phone
    exists, err := s.redis.Exists(ctx, sentKey).Result()
    if err != nil {
        return fmt.Errorf("failed to check sent record: %w", err)
    }
    if exists > 0 {
        return fmt.Errorf("验证码已发送，请 60 秒后再试")
    }
    
    // 2. 检查每小时发送次数限制
    canSend, err := s.CheckRateLimit(ctx, phone)
    if err != nil {
        return err
    }
    if !canSend {
        return fmt.Errorf("发送次数过多，请 1 小时后再试")
    }
    
    // 3. 发送验证码
    code, err := s.smsClient.SendVerificationCode(ctx, phone)
    if err != nil {
        return fmt.Errorf("failed to send verification code: %w", err)
    }
    
    // 4. 存储验证码到 Redis（5 分钟过期）
    codeKey := sms.RedisKeyCodePrefix + phone
    if err := s.redis.Set(ctx, codeKey, code, sms.CodeExpiration).Err(); err != nil {
        return fmt.Errorf("failed to store verification code: %w", err)
    }
    
    // 5. 记录发送时间（60 秒过期）
    if err := s.redis.Set(ctx, sentKey, time.Now().Unix(), sms.SentExpiration).Err(); err != nil {
        return fmt.Errorf("failed to store sent record: %w", err)
    }
    
    // 6. 增加发送次数计数
    countKey := fmt.Sprintf("%s%s:%s", sms.RedisKeyCountPrefix, phone, time.Now().Format("2006010215"))
    if err := s.redis.Incr(ctx, countKey).Err(); err != nil {
        return fmt.Errorf("failed to increment send count: %w", err)
    }
    if err := s.redis.Expire(ctx, countKey, sms.CountExpiration).Err(); err != nil {
        return fmt.Errorf("failed to set count expiration: %w", err)
    }
    
    return nil
}

// VerifyCode 验证验证码
func (s *smsService) VerifyCode(ctx context.Context, phone, code string) (bool, error) {
    // 从 Redis 获取存储的验证码
    codeKey := sms.RedisKeyCodePrefix + phone
    storedCode, err := s.redis.Get(ctx, codeKey).Result()
    if err == redis.Nil {
        return false, fmt.Errorf("验证码已过期或不存在")
    }
    if err != nil {
        return false, fmt.Errorf("failed to get verification code: %w", err)
    }
    
    // 比较验证码
    if storedCode != code {
        return false, nil
    }
    
    // 验证成功后删除验证码（防止重复使用）
    if err := s.redis.Del(ctx, codeKey).Err(); err != nil {
        return true, fmt.Errorf("failed to delete verification code: %w", err)
    }
    
    return true, nil
}

// CheckRateLimit 检查发送频率限制
func (s *smsService) CheckRateLimit(ctx context.Context, phone string) (bool, error) {
    // 获取当前小时的发送次数
    countKey := fmt.Sprintf("%s%s:%s", sms.RedisKeyCountPrefix, phone, time.Now().Format("2006010215"))
    count, err := s.redis.Get(ctx, countKey).Int()
    if err == redis.Nil {
        return true, nil // 没有记录，可以发送
    }
    if err != nil {
        return false, fmt.Errorf("failed to get send count: %w", err)
    }
    
    // 检查是否超过限制
    return count < sms.MaxSendCount, nil
}
```

## 3. 创建 SMS Handler

创建 `internal/handler/sms_handler.go`：

```go
package handler

import (
    "net/http"
    "regexp"
    
    "go-shop/internal/service"
    "go-shop/pkg/response"
    
    "github.com/gin-gonic/gin"
)

// SMSHandler 短信处理器
type SMSHandler struct {
    smsService service.SMSService
}

// NewSMSHandler 创建短信处理器
func NewSMSHandler(smsService service.SMSService) *SMSHandler {
    return &SMSHandler{
        smsService: smsService,
    }
}

// SendVerificationCodeRequest 发送验证码请求
type SendVerificationCodeRequest struct {
    Phone string `json:"phone" binding:"required"`
}

// SendVerificationCode 发送验证码
func (h *SMSHandler) SendVerificationCode(c *gin.Context) {
    var req SendVerificationCodeRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "参数错误")
        return
    }
    
    // 验证手机号格式
    if !isValidPhone(req.Phone) {
        response.Error(c, http.StatusBadRequest, "手机号格式不正确")
        return
    }
    
    // 发送验证码
    if err := h.smsService.SendVerificationCode(c.Request.Context(), req.Phone); err != nil {
        response.Error(c, http.StatusInternalServerError, err.Error())
        return
    }
    
    response.Success(c, nil, "验证码已发送")
}

// isValidPhone 验证手机号格式（中国大陆）
func isValidPhone(phone string) bool {
    pattern := `^1[3-9]\d{9}$`
    matched, _ := regexp.MatchString(pattern, phone)
    return matched
}
```

## 4. 注册路由

在 `internal/router/router.go` 中注册路由：

```go
package router

import (
    "go-shop/internal/handler"
    "go-shop/internal/service"
    
    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"
)

func SetupRouter(redis *redis.Client) *gin.Engine {
    r := gin.Default()
    
    // 创建 SMS 服务和处理器
    smsService, err := service.NewSMSService(redis)
    if err != nil {
        log.Printf("Warning: Failed to create SMS service: %v", err)
    } else {
        smsHandler := handler.NewSMSHandler(smsService)
        
        // SMS 路由
        v2 := r.Group("/api/v2")
        {
            v2.POST("/sms/send", smsHandler.SendVerificationCode)
        }
    }
    
    // ... 其他路由
    
    return r
}
```

## 5. 短信验证码登录实现

在 `internal/handler/auth_handler.go` 中添加短信登录：

```go
// SMSLoginRequest 短信登录请求
type SMSLoginRequest struct {
    Phone string `json:"phone" binding:"required"`
    Code  string `json:"code" binding:"required"`
}

// SMSLogin 短信验证码登录
func (h *AuthHandler) SMSLogin(c *gin.Context) {
    var req SMSLoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "参数错误")
        return
    }
    
    // 验证验证码
    valid, err := h.smsService.VerifyCode(c.Request.Context(), req.Phone, req.Code)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, err.Error())
        return
    }
    if !valid {
        response.Error(c, http.StatusBadRequest, "验证码错误")
        return
    }
    
    // 查找或创建用户
    user, err := h.userService.FindOrCreateByPhone(c.Request.Context(), req.Phone)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "登录失败")
        return
    }
    
    // 生成 JWT token
    token, err := h.jwtService.GenerateToken(user.ID)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "生成令牌失败")
        return
    }
    
    response.Success(c, gin.H{
        "token": token,
        "user":  user,
    }, "登录成功")
}
```

## 6. 前端调用示例

```typescript
// 发送验证码
async function sendVerificationCode(phone: string) {
  try {
    const response = await axios.post('/api/v2/sms/send', { phone });
    if (response.data.code === 0) {
      ElMessage.success('验证码已发送');
      // 开始倒计时
      startCountdown();
    } else {
      ElMessage.error(response.data.message);
    }
  } catch (error) {
    ElMessage.error('发送失败，请稍后重试');
  }
}

// 短信验证码登录
async function smsLogin(phone: string, code: string) {
  try {
    const response = await axios.post('/api/v2/auth/login/sms', {
      phone,
      code
    });
    if (response.data.code === 0) {
      // 保存 token
      localStorage.setItem('token', response.data.data.token);
      ElMessage.success('登录成功');
      router.push('/');
    } else {
      ElMessage.error(response.data.message);
    }
  } catch (error) {
    ElMessage.error('登录失败，请稍后重试');
  }
}
```

## 7. 测试

### 单元测试

```go
func TestSMSService_SendVerificationCode(t *testing.T) {
    // 创建 mock Redis 客户端
    mockRedis := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    
    // 创建 SMS 服务
    smsService, err := service.NewSMSService(mockRedis)
    assert.NoError(t, err)
    
    // 测试发送验证码
    err = smsService.SendVerificationCode(context.Background(), "13800138000")
    assert.NoError(t, err)
}
```

### 集成测试

使用 Postman 或 curl 测试：

```bash
# 发送验证码
curl -X POST http://localhost:8080/api/v2/sms/send \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800138000"}'

# 短信登录
curl -X POST http://localhost:8080/api/v2/auth/login/sms \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800138000","code":"123456"}'
```

## 8. 注意事项

1. **配置阿里云短信服务**：需要在阿里云控制台申请短信签名和模板
2. **保护密钥**：AccessKeySecret 应该存储在环境变量中
3. **频率限制**：实现了 60 秒和每小时 5 次的限制
4. **验证码过期**：验证码 5 分钟后自动过期
5. **防止重复使用**：验证成功后立即删除验证码
6. **错误处理**：适当处理各种错误情况
7. **日志记录**：记录关键操作日志
8. **监控告警**：监控短信发送失败率和费用
