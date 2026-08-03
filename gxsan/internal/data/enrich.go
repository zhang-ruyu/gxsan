package data

import (
	"sync"
	"time"

	"github.com/user/gxsan/internal/model"
)

// EnrichStock 获取股票基础行情，并补全分红数据（最近12个月每股股息、
// 连续分红年数、当前股息率），结果写入缓存。
// CLI 与 GUI 共用这一条数据补全路径，避免重复实现。
// forceRefresh 为 true 时跳过缓存读取，强制重新抓数（用于「随时刷新」）。
// 内部对「行情」与「分红历史」两个独立 HTTP 请求并发发起，单只更快。
func (f *EastMoneyFetcher) EnrichStock(cache *Cache, code string, forceRefresh bool) (*model.Stock, error) {
	if cache != nil && !forceRefresh {
		if cd, err := cache.Get(code); err == nil && cd.Stock != nil {
			return cd.Stock, nil
		}
	}

	var (
		stock     *model.Stock
		stockErr  error
		dividends []model.DividendRecord
		divErr    error
	)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		stock, stockErr = f.GetStock(code)
	}()
	go func() {
		defer wg.Done()
		dividends, divErr = f.GetDividendHistory(code)
	}()
	wg.Wait()

	// 行情是必需的；分红缺失不致命（部分股票可能暂无分红记录）。
	if stockErr != nil {
		return nil, stockErr
	}
	if divErr == nil && len(dividends) > 0 {
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

// maxEnrichConcurrency 并发抓取上限，避免一次性打爆行情接口。
const maxEnrichConcurrency = 8

// EnrichStocks 并发补全多只股票行情+分红数据，返回 code->*model.Stock 映射。
// 多只股票之间并发（最多 maxEnrichConcurrency 只在飞），
// 每只股票内部再并发抓行情与分红（见 EnrichStock）。
// 与 EnrichStock 共享同一套缓存与错误处理语义：抓不到的股票不会出现在结果里。
func (f *EastMoneyFetcher) EnrichStocks(cache *Cache, codes []string, forceRefresh bool) map[string]*model.Stock {
	out := make(map[string]*model.Stock, len(codes))
	if len(codes) == 0 {
		return out
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxEnrichConcurrency)

	for _, code := range codes {
		wg.Add(1)
		go func(c string) {
			defer wg.Done()
			sem <- struct{}{} // 获取并发配额
			defer func() { <-sem }()
			s, err := f.EnrichStock(cache, c, forceRefresh)
			if err == nil {
				mu.Lock()
				out[c] = s
				mu.Unlock()
			}
		}(code)
	}
	wg.Wait()
	return out
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
