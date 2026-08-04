package plan

import (
	"math"
	"testing"
)

func TestCompoundSchedule_Length(t *testing.T) {
	cases := []struct {
		years int
		want  int
	}{
		{0, 0},
		{1, 1},
		{10, 10},
		{100, 100}, // 超长年限：不 panic、长度正确
	}
	for _, c := range cases {
		got := CompoundSchedule(1000, c.years, 0.05)
		if len(got) != c.want {
			t.Errorf("years=%d 时长度=%d, want %d", c.years, len(got), c.want)
		}
	}
}

func TestCompoundSchedule_ZeroRate(t *testing.T) {
	// 年化 0%：每年仅追加定投 12*1000=12000，年股息为 0
	rows := CompoundSchedule(1000, 3, 0)
	for i, r := range rows {
		wantAsset := float64((i + 1) * 12 * 1000)
		if math.Abs(r.TotalAsset-wantAsset) > 1e-9 {
			t.Errorf("第%d年总资产=%v, want %v", r.Year, r.TotalAsset, wantAsset)
		}
		if r.AnnualDividend != 0 || r.MonthlyDividend != 0 {
			t.Errorf("零利率时股息应为0, got=%v/%v", r.AnnualDividend, r.MonthlyDividend)
		}
		if r.Year != i+1 {
			t.Errorf("年份标签错: %d", r.Year)
		}
	}
}

func TestCompoundSchedule_ClosedForm(t *testing.T) {
	// 月定投 1000、年化 6%：12 个月后理论资产 = 1000 * ((1+0.005)^12 - 1)/0.005
	inv := 1000.0
	annual := 0.06
	rm := annual / 12
	want := inv * (math.Pow(1+rm, 12) - 1) / rm

	rows := CompoundSchedule(inv, 1, annual)
	got := rows[0].TotalAsset
	if math.Abs(got-want) > 1e-6 {
		t.Errorf("首年末资产=%v, 闭式解=%v", got, want)
	}
	// 年股息 = 总资产 * 年化率
	if math.Abs(rows[0].AnnualDividend-got*annual) > 1e-6 {
		t.Errorf("年股息不一致: %v vs %v", rows[0].AnnualDividend, got*annual)
	}
	if math.Abs(rows[0].MonthlyDividend-rows[0].AnnualDividend/12) > 1e-9 {
		t.Errorf("月股息应为年股息/12")
	}
}

func TestCompoundSchedule_EdgeNoNaN(t *testing.T) {
	// 负收益率、超长年限：结果应有限、不出现 NaN/Inf
	rows := CompoundSchedule(1000, 50, -0.02)
	for _, r := range rows {
		if math.IsNaN(r.TotalAsset) || math.IsInf(r.TotalAsset, 0) {
			t.Fatalf("出现非法资产值: %v", r.TotalAsset)
		}
		if math.IsNaN(r.AnnualDividend) || math.IsInf(r.AnnualDividend, 0) {
			t.Fatalf("出现非法股息值: %v", r.AnnualDividend)
		}
	}
	// 负利率下资产仍应 > 0（每月定投为正）
	if rows[49].TotalAsset <= 0 {
		t.Errorf("负利率长期下资产应仍为正, got=%v", rows[49].TotalAsset)
	}
}
