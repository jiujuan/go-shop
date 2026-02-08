# OSS 客户端集成示例

本文档展示如何在 Go-Shop 项目中集成和使用 OSS 客户端。

## 1. 在主程序中初始化

在 `cmd/server/main.go` 中初始化 OSS 客户端：

```go
package main

import (
    "log"
    "go-shop/config"
    "go-shop/pkg/oss"
)

func main() {
    // 加载配置
    cfg, err := config.LoadConfig("config/config.yaml")
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }

    // 初始化 OSS 客户端（如果配置了 OSS）
    if cfg.OSS.Endpoint != "" {
        ossConfig := &oss.Config{
            Endpoint:        cfg.OSS.Endpoint,
            AccessKeyID:     cfg.OSS.AccessKeyID,
            AccessKeySecret: cfg.OSS.AccessKeySecret,
            BucketName:      cfg.OSS.BucketName,
            Domain:          cfg.OSS.Domain,
            UseSSL:          cfg.OSS.UseSSL,
            Region:          cfg.OSS.Region,
        }

        if err := oss.Init(ossConfig); err != nil {
            log.Fatal("Failed to initialize OSS client:", err)
        }
        log.Println("OSS client initialized successfully")
    }

    // ... 其他初始化代码
}
```

## 2. 在 Service 层使用

创建上传服务 `internal/service/upload_service.go`：

```go
package service

import (
    "context"
    "fmt"
    "io"
    "go-shop/pkg/oss"
)

type UploadService interface {
    UploadImage(ctx context.Context, file io.Reader, filename string, prefix string) (string, error)
    UploadImages(ctx context.Context, files []io.Reader, filenames []string, prefix string) ([]string, error)
    DeleteImage(ctx context.Context, url string) error
}

type uploadService struct {
    ossClient oss.Client
}

func NewUploadService() (UploadService, error) {
    client, err := oss.GetClient()
    if err != nil {
        return nil, err
    }

    return &uploadService{
        ossClient: client,
    }, nil
}

func (s *uploadService) UploadImage(ctx context.Context, file io.Reader, filename string, prefix string) (string, error) {
    return oss.UploadImage(ctx, s.ossClient, file, filename, prefix)
}

func (s *uploadService) UploadImages(ctx context.Context, files []io.Reader, filenames []string, prefix string) ([]string, error) {
    return oss.UploadImages(ctx, s.ossClient, files, filenames, prefix, 9)
}

func (s *uploadService) DeleteImage(ctx context.Context, url string) error {
    objectName := oss.ExtractObjectName(url)
    return s.ossClient.DeleteFile(ctx, objectName)
}
```

## 3. 在 Handler 层使用

创建上传处理器 `internal/handler/upload_handler.go`：

```go
package handler

import (
    "net/http"
    "go-shop/internal/service"
    "go-shop/pkg/response"
    "github.com/gin-gonic/gin"
)

type UploadHandler struct {
    uploadService service.UploadService
}

func NewUploadHandler(uploadService service.UploadService) *UploadHandler {
    return &UploadHandler{
        uploadService: uploadService,
    }
}

// UploadImage 上传单张图片
// POST /api/v2/upload/image
func (h *UploadHandler) UploadImage(c *gin.Context) {
    // 获取上传的文件
    file, header, err := c.Request.FormFile("file")
    if err != nil {
        response.Error(c, http.StatusBadRequest, "Failed to get file")
        return
    }
    defer file.Close()

    // 获取前缀（可选）
    prefix := c.DefaultPostForm("prefix", "uploads")

    // 上传文件
    url, err := h.uploadService.UploadImage(c.Request.Context(), file, header.Filename, prefix)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, err.Error())
        return
    }

    response.Success(c, gin.H{
        "url": url,
    })
}

// UploadImages 批量上传图片
// POST /api/v2/upload/images
func (h *UploadHandler) UploadImages(c *gin.Context) {
    // 获取上传的文件
    form, err := c.MultipartForm()
    if err != nil {
        response.Error(c, http.StatusBadRequest, "Failed to get files")
        return
    }

    files := form.File["files"]
    if len(files) == 0 {
        response.Error(c, http.StatusBadRequest, "No files uploaded")
        return
    }

    if len(files) > 9 {
        response.Error(c, http.StatusBadRequest, "Maximum 9 files allowed")
        return
    }

    // 获取前缀（可选）
    prefix := c.DefaultPostForm("prefix", "uploads")

    // 准备文件读取器
    readers := make([]io.Reader, len(files))
    filenames := make([]string, len(files))

    for i, fileHeader := range files {
        file, err := fileHeader.Open()
        if err != nil {
            response.Error(c, http.StatusInternalServerError, "Failed to open file")
            return
        }
        defer file.Close()

        readers[i] = file
        filenames[i] = fileHeader.Filename
    }

    // 上传文件
    urls, err := h.uploadService.UploadImages(c.Request.Context(), readers, filenames, prefix)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, err.Error())
        return
    }

    response.Success(c, gin.H{
        "urls": urls,
    })
}
```

## 4. 注册路由

在 `internal/router/router.go` 中注册上传路由：

```go
// 上传路由
uploadHandler := handler.NewUploadHandler(uploadService)
v2.POST("/upload/image", uploadHandler.UploadImage)
v2.POST("/upload/images", uploadHandler.UploadImages)
```

## 5. 在评价和退款功能中使用

### 评价服务中上传图片

```go
// 在 ReviewService 中
func (s *reviewService) CreateReview(ctx context.Context, review *entity.Review, imageFiles []io.Reader, imageFilenames []string) error {
    // 上传评价图片
    if len(imageFiles) > 0 {
        urls, err := oss.UploadImages(ctx, s.ossClient, imageFiles, imageFilenames, "reviews", 9)
        if err != nil {
            return fmt.Errorf("failed to upload review images: %w", err)
        }
        
        // 将 URLs 保存到 review.Images（JSON 字段）
        review.Images = urls
    }

    // 保存评价到数据库
    return s.reviewRepo.Create(ctx, review)
}
```

### 退款服务中上传凭证

```go
// 在 RefundService 中
func (s *refundService) CreateRefund(ctx context.Context, refund *entity.Refund, proofFiles []io.Reader, proofFilenames []string) error {
    // 上传退款凭证图片
    if len(proofFiles) > 0 {
        urls, err := oss.UploadImages(ctx, s.ossClient, proofFiles, proofFilenames, "refunds", 9)
        if err != nil {
            return fmt.Errorf("failed to upload refund proof images: %w", err)
        }
        
        // 将 URLs 保存到 refund.Images（JSON 字段）
        refund.Images = urls
    }

    // 保存退款申请到数据库
    return s.refundRepo.Create(ctx, refund)
}
```

## 6. 前端调用示例

### 上传单张图片

```javascript
async function uploadImage(file) {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('prefix', 'products');

    const response = await fetch('/api/v2/upload/image', {
        method: 'POST',
        body: formData,
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });

    const result = await response.json();
    return result.data.url;
}
```

### 批量上传图片

```javascript
async function uploadImages(files) {
    const formData = new FormData();
    files.forEach(file => {
        formData.append('files', file);
    });
    formData.append('prefix', 'reviews');

    const response = await fetch('/api/v2/upload/images', {
        method: 'POST',
        body: formData,
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });

    const result = await response.json();
    return result.data.urls;
}
```

## 7. 配置 MinIO（开发环境）

使用 Docker 启动 MinIO：

```bash
docker run -d \
  -p 9000:9000 \
  -p 9001:9001 \
  --name minio \
  -e "MINIO_ROOT_USER=minioadmin" \
  -e "MINIO_ROOT_PASSWORD=minioadmin" \
  -v /data/minio:/data \
  minio/minio server /data --console-address ":9001"
```

访问 MinIO 控制台：http://localhost:9001

## 8. 配置阿里云 OSS（生产环境）

在 `config.yaml` 中配置阿里云 OSS：

```yaml
oss:
  endpoint: oss-cn-hangzhou.aliyuncs.com
  access_key_id: YOUR_ACCESS_KEY_ID
  access_key_secret: YOUR_ACCESS_KEY_SECRET
  bucket_name: go-shop-prod
  domain: https://cdn.example.com
  use_ssl: true
  region: cn-hangzhou
```

## 注意事项

1. **安全性**：生产环境中应使用环境变量存储敏感信息（Access Key）
2. **性能**：大文件上传应考虑使用分片上传
3. **错误处理**：上传失败时应有重试机制
4. **清理**：删除数据时应同步删除 OSS 中的文件
5. **CDN**：生产环境建议配置 CDN 加速访问
