package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// PasswordConfig 密码加密配置
type PasswordConfig struct {
	Memory      uint32 // 内存使用量 (KB)
	Iterations  uint32 // 迭代次数
	Parallelism uint8  // 并行度
	SaltLength  uint32 // 盐长度
	KeyLength   uint32 // 密钥长度
}

// DefaultPasswordConfig 默认密码配置
var DefaultPasswordConfig = &PasswordConfig{
	Memory:      64 * 1024, // 64 MB
	Iterations:  3,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

// PasswordManager 密码管理器
type PasswordManager struct {
	config *PasswordConfig
}

// NewPasswordManager 创建密码管理器
func NewPasswordManager(config *PasswordConfig) *PasswordManager {
	if config == nil {
		config = DefaultPasswordConfig
	}
	return &PasswordManager{config: config}
}

// HashPassword 加密密码
func (pm *PasswordManager) HashPassword(password string) (string, error) {
	// 生成随机盐
	salt, err := generateRandomBytes(pm.config.SaltLength)
	if err != nil {
		return "", err
	}

	// 使用 Argon2id 加密
	hash := argon2.IDKey([]byte(password), salt, pm.config.Iterations, pm.config.Memory, pm.config.Parallelism, pm.config.KeyLength)

	// 编码为 base64
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// 格式: $argon2id$v=19$m=65536,t=3,p=2$salt$hash
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, pm.config.Memory, pm.config.Iterations, pm.config.Parallelism, b64Salt, b64Hash)

	return encodedHash, nil
}

// VerifyPassword 验证密码
func (pm *PasswordManager) VerifyPassword(password, encodedHash string) (bool, error) {
	// 解析编码的哈希
	salt, hash, config, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// 使用相同参数重新计算哈希
	otherHash := argon2.IDKey([]byte(password), salt, config.Iterations, config.Memory, config.Parallelism, config.KeyLength)

	// 使用常量时间比较防止时序攻击
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}

// generateRandomBytes 生成随机字节
func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// decodeHash 解码哈希字符串
func decodeHash(encodedHash string) (salt, hash []byte, config *PasswordConfig, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, errors.New("invalid hash format")
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, errors.New("incompatible version of argon2")
	}

	config = &PasswordConfig{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &config.Memory, &config.Iterations, &config.Parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	config.SaltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	config.KeyLength = uint32(len(hash))

	return salt, hash, config, nil
}

// SimpleHashPassword 简单密码加密（用于快速开发）
func SimpleHashPassword(password string) (string, error) {
	pm := NewPasswordManager(nil)
	return pm.HashPassword(password)
}

// SimpleVerifyPassword 简单密码验证（用于快速开发）
func SimpleVerifyPassword(password, hash string) (bool, error) {
	pm := NewPasswordManager(nil)
	return pm.VerifyPassword(password, hash)
}