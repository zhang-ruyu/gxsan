package data

import "github.com/user/gxsan/internal/model"

// Fetcher 数据获取接口
type Fetcher interface {
	GetStock(code string) (*model.Stock, error)
	GetDividendHistory(code string) ([]model.DividendRecord, error)
	SearchStock(keyword string) ([]model.StockInfo, error)
}
