package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/user/gxsan/internal/model"
)

// Cache 数据缓存
type Cache struct {
	DataDir    string
	Expiration time.Duration
}

// CacheData 缓存数据结构
type CacheData struct {
	Stock     *model.Stock     `json:"stock"`
	Dividends []model.DividendRecord `json:"dividends"`
	Timestamp time.Time        `json:"timestamp"`
}

// NewCache 创建缓存
func NewCache(dataDir string) *Cache {
	return &Cache{
		DataDir:    dataDir,
		Expiration: 24 * time.Hour, // 缓存24小时
	}
}

// getCachePath 获取缓存文件路径
func (c *Cache) getCachePath(code string) string {
	return filepath.Join(c.DataDir, code+".json")
}

// Get 获取缓存
func (c *Cache) Get(code string) (*CacheData, error) {
	path := c.getCachePath(code)
	
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("缓存不存在")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取缓存失败: %w", err)
	}

	var cacheData CacheData
	if err := json.Unmarshal(data, &cacheData); err != nil {
		return nil, fmt.Errorf("解析缓存失败: %w", err)
	}

	// 检查是否过期
	if time.Since(cacheData.Timestamp) > c.Expiration {
		return nil, fmt.Errorf("缓存已过期")
	}

	return &cacheData, nil
}

// Set 设置缓存
func (c *Cache) Set(code string, stock *model.Stock, dividends []model.DividendRecord) error {
	cacheData := CacheData{
		Stock:     stock,
		Dividends: dividends,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(cacheData)
	if err != nil {
		return fmt.Errorf("序列化缓存失败: %w", err)
	}

	path := c.getCachePath(code)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入缓存失败: %w", err)
	}

	return nil
}

// Clear 清除缓存
func (c *Cache) Clear() error {
	files, err := filepath.Glob(filepath.Join(c.DataDir, "*.json"))
	if err != nil {
		return err
	}

	for _, f := range files {
		os.Remove(f)
	}

	return nil
}
