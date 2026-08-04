package model

// InvestPool 投资池
type InvestPool struct {
	AvailableFund  float64            `yaml:"available_fund"`
	Holdings       map[string]Holding `yaml:"-"`
	TotalAsset     float64            `yaml:"-"`
	MaxPosition    float64            `yaml:"-"`
	MaxPositionPct float64            `yaml:"max_position_pct"`
}

// FundConfig 资金配置
type FundConfig struct {
	AvailableFund  float64 `yaml:"available_fund"`
	MaxPositionPct float64 `yaml:"max_position_pct"`
}

// Recommend 操作推荐
type Recommend struct {
	Stock      Stock
	Action     string
	Amount     float64
	Reason     string
	Priority   int
	Constraint string
}

// CalculatePool 计算投资池
func (p *InvestPool) CalculatePool() {
	p.TotalAsset = p.AvailableFund
	for _, h := range p.Holdings {
		p.TotalAsset += h.MarketValue
	}
	p.MaxPosition = p.TotalAsset * p.MaxPositionPct / 100
}

// GetHoldingRatio 获取持仓占比
func (p *InvestPool) GetHoldingRatio(code string) float64 {
	if p.TotalAsset == 0 {
		return 0
	}
	h, ok := p.Holdings[code]
	if !ok {
		return 0
	}
	return h.MarketValue / p.TotalAsset * 100
}
