package data

import (
	"time"

	"github.com/user/gxsan/internal/model"
)

// EnrichStock 获取股票基础行情，并补全分红数据（最近12个月每股股息、
// 连续分红年数、当前股息率），结果写入缓存。
// CLI 与 GUI 共用这一条数据补全路径，避免重复实现。
// forceRefresh 为 true 时跳过缓存读取，强制重新抓数（用于「随时刷新」）。
func (f *EastMoneyFetcher) EnrichStock(cache *Cache, code string, forceRefresh bool) (*model.Stock, error) {
	if cache != nil && !forceRefresh {
		if cd, err := cache.Get(code); err == nil && cd.Stock != nil {
			return cd.Stock, nil
		}
	}

	stock, err := f.GetStock(code)
	if err != nil {
		return nil, err
	}

	dividends, err := f.GetDividendHistory(code)
	if err == nil && len(dividends) > 0 {
		stock.DividendPerShare = annualDividend(dividends)
		stock.LastDivDate = dividends[0].Date
		if len(dividends) > 1 {
			stock.NextDivDate = dividends[1].Date
		}
	}

	stock.DividendYield = CalculateDividendYield(stock.Price, stock.DividendPerShare)
	stock.DividendYears = dividendYears(dividends)

	if cache != nil {
		_ = cache.Set(code, stock, dividends)
	}

	return stock, nil
}

// annualDividend 计算最近12个月内的每股股息总额
func annualDividend(dividends []model.DividendRecord) float64 {
	if len(dividends) == 0 {
		return 0
	}

	now := time.Now()
	oneYearAgo := now.AddDate(0, -12, 0)
	total := 0.0

	for _, div := range dividends {
		divDate, err := time.Parse("2006-01-02", div.Date)
		if err != nil {
			continue
		}
		if divDate.After(oneYearAgo) && divDate.Before(now) {
			total += div.Amount
		}
	}

	return total
}

// dividendYears 计算连续分红年数
func dividendYears(dividends []model.DividendRecord) int {
	if len(dividends) == 0 {
		return 0
	}

	years := 1
	for i := 1; i < len(dividends); i++ {
		t1, err1 := time.Parse("2006-01-02", dividends[i-1].Date)
		t2, err2 := time.Parse("2006-01-02", dividends[i].Date)
		if err1 != nil || err2 != nil {
			continue
		}

		yearDiff := t1.Year() - t2.Year()
		if yearDiff == 1 || yearDiff == 0 {
			years++
		} else {
			break
		}
	}

	return years
}
