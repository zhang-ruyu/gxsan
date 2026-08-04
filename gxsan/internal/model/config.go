package model

// Config 全局配置
type Config struct {
	DefaultTargetYield float64       `yaml:"default_target_yield"`
	MinDividendYears   int           `yaml:"min_dividend_years"`
	CheapDiscount      float64       `yaml:"cheap_discount"`
	ExpensivePremium   float64       `yaml:"expensive_premium"`
	LifecycleStage     int           `yaml:"lifecycle_stage"` // 1启动 2滚雪球 3自由 4收获
	TargetAnnualDividend float64     `yaml:"target_annual_dividend"` // 养老目标：目标年分红（元），用于退休金进度条
	Fund               FundConfig    `yaml:"fund"`
	Watchlist          []StockConfig `yaml:"watchlist"`
	Portfolio          []Holding     `yaml:"portfolio"`
}

// FindStock 在监控列表中查找指定代码的股票配置，找不到返回 nil
func (c *Config) FindStock(code string) *StockConfig {
	for i := range c.Watchlist {
		if c.Watchlist[i].Code == code {
			return &c.Watchlist[i]
		}
	}
	return nil
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		DefaultTargetYield: 4.0,
		MinDividendYears:   3,
		CheapDiscount:      0.8,
		ExpensivePremium:   1.2,
		LifecycleStage:     2,
		TargetAnnualDividend: 0,
		Fund: FundConfig{
			AvailableFund:  0,
			MaxPositionPct: 30,
		},
		Watchlist: []StockConfig{},
		Portfolio: []Holding{},
	}
}
