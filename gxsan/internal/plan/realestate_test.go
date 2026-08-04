package plan

import (
	"math"
	"testing"
)

func TestRealEstateToEquity_Basic(t *testing.T) {
	// 本金 100万、租售比 2%、股息率 5% -> 租金 2万、股息 5万、倍数 2.5
	c := RealEstateToEquity(1_000_000, 0.02, 0.05)
	if math.Abs(c.RealEstateIncome-20_000) > 1e-6 {
		t.Errorf("年租金=%v, want 20000", c.RealEstateIncome)
	}
	if math.Abs(c.EquityIncome-50_000) > 1e-6 {
		t.Errorf("年股息=%v, want 50000", c.EquityIncome)
	}
	if math.Abs(c.IncomeMultiple-2.5) > 1e-9 {
		t.Errorf("现金流倍数=%v, want 2.5", c.IncomeMultiple)
	}
	if c.Principal != 1_000_000 {
		t.Errorf("本金未透传")
	}
}

func TestRealEstateToEquity_ZeroRealEstateYield(t *testing.T) {
	// 租售比 0：避免除零，倍数应为 0（而非 NaN/Inf）
	c := RealEstateToEquity(1_000_000, 0, 0.05)
	if c.RealEstateIncome != 0 {
		t.Errorf("年租金应为0, got=%v", c.RealEstateIncome)
	}
	if math.Abs(c.EquityIncome-50_000) > 1e-6 {
		t.Errorf("股权收入应正常: %v", c.EquityIncome)
	}
	if c.IncomeMultiple != 0 {
		t.Errorf("租售比0时倍数应为0, got=%v", c.IncomeMultiple)
	}
	if math.IsNaN(c.IncomeMultiple) || math.IsInf(c.IncomeMultiple, 0) {
		t.Errorf("不应产生非法倍数: %v", c.IncomeMultiple)
	}
}

func TestRealEstateToEquity_ZeroPrincipal(t *testing.T) {
	c := RealEstateToEquity(0, 0.02, 0.05)
	if c.RealEstateIncome != 0 || c.EquityIncome != 0 || c.IncomeMultiple != 0 {
		t.Errorf("本金0时各项应均为0, got=%+v", c)
	}
}

func TestRealEstateToEquity_NegativeYield(t *testing.T) {
	// 负收益率不应 panic；因 re<=0 触发倍数防护，IncomeMultiple 返回 0（而非 NaN/Inf）
	c := RealEstateToEquity(1_000_000, -0.02, -0.05)
	if math.IsNaN(c.IncomeMultiple) || math.IsInf(c.IncomeMultiple, 0) {
		t.Fatalf("负收益率不应产生非法倍数: %v", c.IncomeMultiple)
	}
	if c.IncomeMultiple != 0 {
		t.Errorf("re<=0 时倍数应防护为 0, got=%v", c.IncomeMultiple)
	}
	// 负租金/负股息本身应有限数值
	if math.IsNaN(c.RealEstateIncome) || math.IsInf(c.RealEstateIncome, 0) {
		t.Errorf("负租金不应非法: %v", c.RealEstateIncome)
	}
	if math.IsNaN(c.EquityIncome) || math.IsInf(c.EquityIncome, 0) {
		t.Errorf("负股息不应非法: %v", c.EquityIncome)
	}
}
