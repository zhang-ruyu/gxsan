package model

// Stock 股票实时数据
type Stock struct {
	Code             string  `json:"code"`
	Name             string  `json:"name"`
	Price            float64 `json:"price"`
	Change           float64 `json:"change"`
	PE               float64 `json:"pe"`
	PB               float64 `json:"pb"`
	DividendPerShare float64 `json:"dividend_per_share"`
	DividendYield    float64 `json:"dividend_yield"`
	DividendYears    int     `json:"dividend_years"`
	LastDivDate      string  `json:"last_div_date"`
	NextDivDate      string  `json:"next_div_date"`
}

// StockConfig 股票配置
type StockConfig struct {
	Code         string      `yaml:"code"`
	Name         string      `yaml:"name"`
	TargetYield  float64     `yaml:"target_yield"`
	GridStrategy GridStrategy `yaml:"grid_strategy"`
}

// GridStrategy 网格策略
type GridStrategy struct {
	BuyGrids  [5]GridLevel `yaml:"buy_grids"`
	SellGrids [5]GridLevel `yaml:"sell_grids"`
}

// GridLevel 网格档位
type GridLevel struct {
	Yield  float64 `yaml:"yield"`
	Amount float64 `yaml:"amount"`
}

// DividendRecord 分红记录
type DividendRecord struct {
	Date       string  `json:"date"`
	Amount     float64 `json:"amount"`
	ExDivDate  string  `json:"ex_div_date"`
	PayDate    string  `json:"pay_date"`
}

// StockInfo 股票搜索结果
type StockInfo struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// Holding 持仓
type Holding struct {
	Code         string  `yaml:"code"`
	Name         string  `yaml:"name"`
	Shares       int     `yaml:"shares"`
	AvgCost      float64 `yaml:"avg_cost"`
	TotalCost    float64 `yaml:"total_cost"` // 真实投入总额（含费用），用于一键修正推导精确每股成本
	CurrentPrice float64 `yaml:"current_price,omitempty"`
	MarketValue  float64 `json:"market_value,omitempty"`
	Dividend     float64 `yaml:"dividend,omitempty"`
	YieldOnCost  float64 `yaml:"yield_on_cost,omitempty"`
}
