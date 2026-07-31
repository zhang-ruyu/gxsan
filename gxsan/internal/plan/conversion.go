package plan

import (
	"fmt"

	"github.com/user/gxsan/internal/model"
)

// StockCompare 个股对比项（对比框架的可量化维度）
type StockCompare struct {
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	CurrentYield  float64 `json:"current_yield"`
	PE            float64 `json:"pe"`
	PB            float64 `json:"pb"`
	DividendYears int     `json:"dividend_years"`
	YieldOnCost   float64 `json:"yield_on_cost"`
}

// CompareResult 个股对比结果
type CompareResult struct {
	A      StockCompare `json:"a"`
	B      StockCompare `json:"b"`
	Better string      `json:"better"` // "A" / "B" / "持平"
	Reason string      `json:"reason"`
}

// CompareStocks 对比两只股票（基于当前股息率/估值/分红稳定性）
// ayoc/byoc 为各自持仓的成本股息率（无持仓传0）。
// 对应技能「标的比较框架」：当前股息率、估值(PB)、分红稳定性(年数) 等维度。
func CompareStocks(a, b *model.Stock, ayoc, byoc float64) CompareResult {
	ca := StockCompare{Code: a.Code, Name: a.Name, CurrentYield: a.DividendYield, PE: a.PE, PB: a.PB, DividendYears: a.DividendYears, YieldOnCost: ayoc}
	cb := StockCompare{Code: b.Code, Name: b.Name, CurrentYield: b.DividendYield, PE: b.PE, PB: b.PB, DividendYears: b.DividendYears, YieldOnCost: byoc}

	better := "持平"
	reason := "两者股息率相近"
	if a.DividendYield > b.DividendYield+0.5 {
		better = "A"
		reason = fmt.Sprintf("%s 当前股息率 %.2f%% 高于 %s 的 %.2f%%", a.Name, a.DividendYield, b.Name, b.DividendYield)
	} else if b.DividendYield > a.DividendYield+0.5 {
		better = "B"
		reason = fmt.Sprintf("%s 当前股息率 %.2f%% 高于 %s 的 %.2f%%", b.Name, b.DividendYield, a.Name, a.DividendYield)
	}
	return CompareResult{A: ca, B: cb, Better: better, Reason: reason}
}

// AssetConversion 房产转股权现金流对比
type AssetConversion struct {
	Principal        float64 `json:"principal"`          // 本金
	RealEstateYield  float64 `json:"real_estate_yield"`  // 租售比(年化, 如0.02)
	EquityYield      float64 `json:"equity_yield"`       // 红利股息率(如0.05)
	RealEstateIncome float64 `json:"real_estate_income"` // 年租金
	EquityIncome     float64 `json:"equity_income"`      // 年股息
	IncomeMultiple   float64 `json:"income_multiple"`    // 现金流倍数(股权/房产)
}

// RealEstateToEquity 房产→红利股权现金流对比（对应技能「资产转换策略」）
// 老破小租售比1-2%、流动性差、需维护；红利资产股息率4-5%+、T+1变现、零维护。
func RealEstateToEquity(principal, realEstateYield, equityYield float64) AssetConversion {
	re := principal * realEstateYield
	eq := principal * equityYield
	mult := 0.0
	if re > 0 {
		mult = eq / re
	}
	return AssetConversion{
		Principal:        principal,
		RealEstateYield:  realEstateYield,
		EquityYield:      equityYield,
		RealEstateIncome: re,
		EquityIncome:     eq,
		IncomeMultiple:   mult,
	}
}
