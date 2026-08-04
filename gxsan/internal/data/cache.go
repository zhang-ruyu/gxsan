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
	Expiration time.Duration // 兜底上限；实际由 marketCacheTTL() 动态决定
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
		Expiration: 8 * time.Hour, // 兜底上限（非交易时段上限）
	}
}

// marketCacheTTL 返回当前时段适用的缓存有效期。
// 交易时段（工作日 9:25–15:00 北京时间）3 分钟刷新一次，
// 获取准实时价；收盘后股价不变，8 小时 TTL 足以覆盖到次日开盘前。
// 不依赖交易日历（无法区分节假日），但假期内刷新也只是多请求一次
// 返回相同的最近收盘价，无副作用。
func marketCacheTTL() time.Duration {
	now := time.Now()
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		return 8 * time.Hour
	}
	h, m := now.Hour(), now.Minute()
	// 9:25–15:00（含午间休市 11:30–13:00，价格虽不变但 3 分钟 TTL 浪费可忽略）
	if (h == 9 && m >= 25) || (h >= 10 && h < 15) {
		return 3 * time.Minute
	}
	return 8 * time.Hour
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

	// 检查是否过期（交易时段 3 分钟，非交易时段 8 小时）
	ttl := marketCacheTTL()
	if c.Expiration > 0 && c.Expiration < ttl {
		ttl = c.Expiration // 允许外部覆盖更短的 TTL
	}
	if time.Since(cacheData.Timestamp) > ttl {
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
