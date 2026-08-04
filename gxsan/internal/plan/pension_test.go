package plan

import (
	"math"
	"testing"
)

// round3 模拟前端 toFixed(3) 的精度契约
func round3(x float64) float64 { return math.Round(x*1e3) / 1e3 }

func TestRequiredPrincipal(t *testing.T) {
	cases := []struct {
		name     string
		target   float64
		costYield float64
		want     float64
	}{
		{"正常 5%", 10000, 0.05, 200000},
		{"正常 7% 非整除", 1000, 0.07, 14285.714285714286},
		{"成本股息率为 0 -> 返回 0（避免除零 NaN）", 10000, 0, 0},
		{"成本股息率为负 -> 返回 0", 10000, -0.03, 0},
		{"目标为 0 -> 本金 0", 0, 0.05, 0},
		{"极小股息率 -> 极大但有限", 100, 1e-9, 1e11},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := RequiredPrincipal(c.target, c.costYield)
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Fatalf("结果非法: got=%v", got)
			}
			if math.Abs(got-c.want) > 1e-6 {
				t.Errorf("RequiredPrincipal(%v,%v)=%v, want %v", c.target, c.costYield, got, c.want)
			}
		})
	}
}

func TestRequiredPrincipal_3DecimalContract(t *testing.T) {
	// 与前端 toFixed(3) 契约一致：分红 1000、成本股息率 7% -> 14285.714
	got := round3(RequiredPrincipal(1000, 0.07))
	want := 14285.714
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("3位小数契约不符: got=%v want=%v", got, want)
	}
	// 整除场景应保持精确
	if round3(RequiredPrincipal(10000, 0.05)) != 200000.000 {
		t.Errorf("整除场景精度异常")
	}
}

func TestPensionPlan(t *testing.T) {
	// 月分红 1000 -> 年 12000；成本股息率 5/6/7% -> 240000/200000/171428.571
	rows := PensionPlan(1000, []float64{0.05, 0.06, 0.07})
	if len(rows) != 3 {
		t.Fatalf("行数=%d, want 3", len(rows))
	}
	wantPrincipal := []float64{240000, 200000, 171428.57142857142}
	for i, w := range wantPrincipal {
		if math.Abs(rows[i].RequiredPrincipal-w) > 1e-6 {
			t.Errorf("rows[%d].RequiredPrincipal=%v, want %v", i, rows[i].RequiredPrincipal, w)
		}
		if rows[i].CostYield != []float64{0.05, 0.06, 0.07}[i] {
			t.Errorf("rows[%d].CostYield=%v", i, rows[i].CostYield)
		}
	}

	// 含股息率=0 的边界：应安全返回 0，不 panic/NaN
	rows2 := PensionPlan(1000, []float64{0, 0.05})
	if rows2[0].RequiredPrincipal != 0 {
		t.Errorf("股息率=0 时应返回 0, got=%v", rows2[0].RequiredPrincipal)
	}
	if math.IsNaN(rows2[0].RequiredPrincipal) {
		t.Errorf("股息率=0 不应产生 NaN")
	}
}

func TestRetirementNote(t *testing.T) {
	cases := map[int]string{
		1: "启动期：坚持定投、养成复投习惯，不急于看效果。",
		2: "滚雪球期：坚决不提现，分红100%复投，下跌加仓。",
		3: "自由期：部分消费+部分复投，克制消费冲动。",
		4: "收获期：提取改善生活，其余复投，考虑传承；提取率建议≤5%。",
	}
	for stage, want := range cases {
		if got := RetirementNote(stage); got != want {
			t.Errorf("RetirementNote(%d)=%q, want %q", stage, got, want)
		}
	}
	// 默认分支
	if got := RetirementNote(99); got != "滚雪球期：坚决不提现，分红100%复投。" {
		t.Errorf("默认分支异常: %q", got)
	}
}
