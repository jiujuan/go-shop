# OSS Package

OSS (Object Storage Service) 包提供了对象存储服务的封装，支持 MinIO 和 S3 兼容的存储服务。

## 功能特性

- 文件上传、下载、删除
- 图片格式和大小验证
- 批量上传支持
- 自动生成唯一文件名
- CDN 域名支持
- 线程安全的全局客户端

## 配置

在 `config.yaml` 中配置 OSS：

```yaml
oss:
  endpoint: localhost:9000        # OSS 端点地址
  access_key_id: minioadmin       # Access Key ID
  access_key_secret: minioadmin   # Access Key Secret
  bucket_name: go-shop            # 存储桶名称
  domain: ""                      # CDN 域名（可选）
  use_ssl: false                  # 是否使用 SSL
  region: ""                      # 区域（可选）
```

## 使用示例

### 初始化

```go
import "go-shop/pkg/oss"

// 初始化 OSS 客户端
config := &oss.Config{
    Endpoint:        "localhost:9000",
    AccessKeyID:     "minioadmin",
    AccessKeySecret: "minioadmin",
    BucketName:      "go-shop",
    UseSSL:          false,
}

err := oss.Init(config)
if err != nil {
    log.Fatal(err)
}
```

### 上传文件

```go
client := oss.MustGetClient()

// 上传单个文件
file, _ := os.Open("image.jpg")
defer file.Close()

url, err := oss.UploadImage(ctx, client, file, "image.jpg", "products")
if err != nil {
    log.Fatal(err)
}

fmt.Println("File URL:", url)
```

### 批量上传

```go
files := []io.Reader{file1, file2, file3}
filenames := []string{"image1.jpg", "image2.jpg", "image3.jpg"}

urls, err := oss.UploadImages(ctx, client, files, filenames, "reviews", 9)
if err != nil {
    log.Fatal(err)
}
```

### 下载文件

```go
reader, err := client.DownloadFile(ctx, "products/2024/01/02/uuid.jpg")
if err != nil {
    log.Fatal(err)
}
defer reader.Close()

// 读取文件内容
data, _ := io.ReadAll(reader)
```

### 删除文件

```go
err := client.DeleteFile(ctx, "products/2024/01/02/uuid.jpg")
if err != nil {
    log.Fatal(err)
}
```

## 图片验证

支持的图片格式：
- JPEG (.jpg, .jpeg)
- PNG (.png)
- GIF (.gif)

最大文件大小：5MB

## 文件命名规则

自动生成的文件名格式：`{prefix}/{date}/{uuid}.{ext}`

示例：`products/2024/01/02/550e8400-e29b-41d4-a716-446655440000.jpg`

## 注意事项

1. 确保 MinIO 或 S3 服务已启动并可访问
2. 存储桶不存在时会自动创建
3. 上传前会自动验证图片格式和大小
4. 使用 CDN 域名时，返回的 URL 会使用 CDN 地址
5. 所有操作都支持 context 超时控制

## 测试

```bash
# 运行单元测试
go test ./pkg/oss/...

# 运行测试并显示覆盖率
go test -cover ./pkg/oss/...
```

## 依赖

- github.com/minio/minio-go/v7 - MinIO Go SDK
- github.com/google/uuid - UUID 生成
