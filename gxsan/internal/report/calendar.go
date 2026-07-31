package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/user/gxsan/internal/model"
)

// GenerateCalendar 生成股利日历
func (r *Reporter) GenerateCalendar(days int) (string, error) {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("══════════════════════════════════════════════════════════════════\n"))
	result.WriteString(fmt.Sprintf("                    股利日历 - 未来%d天\n", days))
	result.WriteString(fmt.Sprintf("══════════════════════════════════════════════════════════════════\n\n"))

	now := time.Now()

	// 获取持仓的分红信息
	result.WriteString(fmt.Sprintf("  ── 持仓股息预估 ─────────────────────────────────────────────┐\n"))

	totalAnnualDividend := 0.0

	for _, holding := range r.config.Portfolio {
		stock, err := r.fetcher.EnrichStock(r.cache, holding.Code, false)
		if err != nil {
			continue
		}

		// 计算持仓的年化股息
		annualDividend := stock.DividendPerShare * float64(holding.Shares)
		totalAnnualDividend += annualDividend

		// 基于历史分红日期估算下次分红时间
		nextDivEstimate := r.estimateNextDividendDate(stock)
		daysUntil := 0
		if !nextDivEstimate.IsZero() {
			daysUntil = int(nextDivDateSub(nextDivEstimate, now))
		}

		status := "待公布"
		if daysUntil > 0 && daysUntil <= days {
			status = fmt.Sprintf("约%d天后", daysUntil)
		} else if daysUntil > days {
			status = fmt.Sprintf("约%d天后", daysUntil)
		}

		result.WriteString(fmt.Sprintf("  │ %s (%s)  每股¥%.2f  年化¥%.0f  %s\n",
			stock.Name, stock.Code, stock.DividendPerShare, annualDividend, status))
	}

	result.WriteString(fmt.Sprintf("  └────────────────────────────────────────────────────────────┘\n\n"))

	// 分红历史记录
	result.WriteString(fmt.Sprintf("  ── 近期分红记录 ─────────────────────────────────────────────┐\n"))
	result.WriteString(fmt.Sprintf("  日期        代码    名称      每股股息  持仓股息\n"))
	result.WriteString(fmt.Sprintf("  ────────────────────────────────────────────────────────────────\n"))

	for _, holding := range r.config.Portfolio {
		stock, err := r.fetcher.EnrichStock(r.cache, holding.Code, false)
		if err != nil {
			continue
		}

		// 从缓存获取分红历史
		cacheData, err := r.cache.Get(holding.Code)
		if err == nil && len(cacheData.Dividends) > 0 {
			// 显示最近3次分红记录
			count := 0
			for _, div := range cacheData.Dividends {
				if count >= 3 {
					break
				}
				dividend := div.Amount * float64(holding.Shares)
				result.WriteString(fmt.Sprintf("  %s  %s  %s  ¥%.2f    ¥%.0f\n",
					div.Date, stock.Code, stock.Name, div.Amount, dividend))
				count++
			}
		}
	}

	result.WriteString(fmt.Sprintf("  ────────────────────────────────────────────────────────────────\n\n"))

	// 统计信息
	result.WriteString(fmt.Sprintf("  持仓数量:     %d只\n", len(r.config.Portfolio)))
	result.WriteString(fmt.Sprintf("  年化股息总额: ¥%.0f\n", totalAnnualDividend))
	result.WriteString(fmt.Sprintf("  月均股息:     ¥%.0f\n", totalAnnualDividend/12))
	result.WriteString(fmt.Sprintf("══════════════════════════════════════════════════════════════════\n"))

	return result.String(), nil
}

// nextDivDateSub 计算时间差（天）
func nextDivDateSub(a, b time.Time) float64 {
	return a.Sub(b).Hours() / 24
}

// estimateNextDividendDate 估算下次分红日期
// 基于历史分红模式粗略估算：年报分红约在7月，中报分红约在12月。
func (r *Reporter) estimateNextDividendDate(stock *model.Stock) time.Time {
	now := time.Now()
	year := now.Year()

	estimatedDates := []time.Time{
		time.Date(year, 7, 15, 0, 0, 0, 0, time.Local),   // 年报分红
		time.Date(year, 12, 15, 0, 0, 0, 0, time.Local),  // 中报分红
		time.Date(year+1, 7, 15, 0, 0, 0, 0, time.Local), // 明年年报分红
	}

	// 找到最近的未来日期
	for _, d := range estimatedDates {
		if d.After(now) {
			return d
		}
	}

	return time.Time{}
}
