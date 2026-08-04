package plan

import (
	"testing"

	"github.com/user/gxsan/internal/model"
)

func TestCompareStocks_AWins(t *testing.T) {
	a := &model.Stock{Code: "600900", Name: "长江电力", DividendYield: 3.5, PE: 18, PB: 2.1, DividendYears: 20}
	b := &model.Stock{Code: "601398", Name: "工商银行", DividendYield: 2.8, PE: 6, PB: 0.6, DividendYears: 18}
	r := CompareStocks(a, b, 4.0, 3.0)
	if r.Better != "A" {
		t.Errorf("Better=%q, want A (长江电力股息率高 0.7%%>0.5%%)", r.Better)
	}
	if r.A.YieldOnCost != 4.0 || r.B.YieldOnCost != 3.0 {
		t.Errorf("成本股息率未透传: A=%v B=%v", r.A.YieldOnCost, r.B.YieldOnCost)
	}
	if r.A.Code != "600900" || r.B.Code != "601398" {
		t.Errorf("代码未透传")
	}
}

func TestCompareStocks_BWins(t *testing.T) {
	a := &model.Stock{Code: "600900", Name: "长江电力", DividendYield: 2.8, PE: 18, PB: 2.1, DividendYears: 20}
	b := &model.Stock{Code: "601398", Name: "工商银行", DividendYield: 3.5, PE: 6, PB: 0.6, DividendYears: 18}
	r := CompareStocks(a, b, 3.0, 4.0)
	if r.Better != "B" {
		t.Errorf("Better=%q, want B", r.Better)
	}
}

func TestCompareStocks_Tie(t *testing.T) {
	a := &model.Stock{Code: "A", Name: "甲", DividendYield: 3.0}
	b := &model.Stock{Code: "B", Name: "乙", DividendYield: 3.2} // 差 0.2% < 0.5% 阈值
	r := CompareStocks(a, b, 0, 0)
	if r.Better != "持平" {
		t.Errorf("Better=%q, want 持平 (差 0.2%%<0.5%%)", r.Better)
	}
}

func TestCompareStocks_ZeroYoc(t *testing.T) {
	// 无持仓传 0：YieldOnCost 应为 0，不 panic
	a := &model.Stock{Code: "A", Name: "甲", DividendYield: 3.0}
	b := &model.Stock{Code: "B", Name: "乙", DividendYield: 2.5}
	r := CompareStocks(a, b, 0, 0)
	if r.A.YieldOnCost != 0 || r.B.YieldOnCost != 0 {
		t.Errorf("零成本股息率透传异常")
	}
}
