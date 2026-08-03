package model

// 以下结构体为 GUI（Wails 前端）与 CLI 共用的视图模型（DTO）。
// 原先分散定义在 gxsan/app.go 中，现统一收敛到 model 包，避免重复定义。

// PortfolioItem 持仓项（GUI 展示用）
type PortfolioItem struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Account     string  `json:"account"` // 所属账户（三账户体系）
	Shares      int     `json:"shares"`
	AvgCost     float64 `json:"avg_cost"`
	Price       float64 `json:"price"`
	MarketValue float64 `json:"market_value"`
	Profit      float64 `json:"profit"`
	ProfitPct   float64 `json:"profit_pct"`
	YieldOnCost float64 `json:"yield_on_cost"` // 成本股息率 YoC
}

// FundInfo 资金信息（GUI 展示用）
type FundInfo struct {
	AvailableFund  float64 `json:"available_fund"`
	TotalMarket    float64 `json:"total_market"`
	TotalAssets    float64 `json:"total_assets"`
	MaxPosition    float64 `json:"max_position"`
	MaxPositionPct float64 `json:"max_position_pct"`
}

// WatchlistItem 监控项（GUI 展示用）
type WatchlistItem struct {
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	TargetYield   float64 `json:"target_yield"`
	Price         float64 `json:"price"`
	DividendYield float64 `json:"dividend_yield"`
	Signal        string  `json:"signal"`
}

// GridStrategyInfo 网格策略信息（GUI 展示用）
type GridStrategyInfo struct {
	BuyGrids  []GridLevel `json:"buy_grids"`
	SellGrids []GridLevel `json:"sell_grids"`
}

// DividendHistoryItem 单年分红历史（详情页用）
type DividendHistoryItem struct {
	Year             int     `json:"year"`
	DividendPerShare float64 `json:"dividend_per_share"`
	Yield            float64 `json:"yield"`
}

// StockDetail 单只股票详情（GUI Detail.vue JSON 契约）
// 字段名需与前端 Detail.vue 读取的 report.data.* 保持一致。
type StockDetail struct {
	StockCode           string              `json:"stock_code"`
	StockName           string              `json:"stock_name"`
	CurrentPrice        float64             `json:"current_price"`
	Change              float64             `json:"change"`
	PE                  float64             `json:"pe"`
	PB                  float64             `json:"pb"`
	LatestDividendYield float64             `json:"latest_dividend_yield"`
	AvgDividendYield   float64             `json:"avg_dividend_yield"` // 近3年平均静态股息率
	TargetYield        float64             `json:"target_yield"`
	DividendPerShare   float64             `json:"dividend_per_share"`
	DividendYears      int                 `json:"dividend_years"`
	YieldOnCost        float64             `json:"yield_on_cost"` // 成本股息率（有持仓时）
	Account             string              `json:"account"`       // 所属账户（有持仓时）
	Signal              string             `json:"signal"`
	Valuation           string             `json:"valuation"`
	Score               int                `json:"score"`
	Reason              string             `json:"reason"`
	CheapPrice          float64             `json:"cheap_price"`
	FairPrice           float64             `json:"fair_price"`
	ExpensivePrice      float64             `json:"expensive_price"`
	CheapDiscount       float64             `json:"cheap_discount"`
	ExpensivePremium    float64             `json:"expensive_premium"`
	GridStrategy        *GridStrategyInfo   `json:"grid_strategy"`
	History             []DividendHistoryItem `json:"history"`
}

// DividendSummary 分红按年/账户汇总（GUI 分红汇总页用）
// 按自然年汇总各持仓分红金额、按账户分组汇总，便于评估分红现金流。
type DividendSummary struct {
	ByAccount           []AccountDividend `json:"by_account"`
	ByYear              []YearDividend    `json:"by_year"`
	TotalAnnualDividend float64           `json:"total_annual_dividend"` // 当前持仓年化分红总额
	TotalMarketValue    float64           `json:"total_market_value"`
	TotalHoldings       int               `json:"total_holdings"`
}

// AccountDividend 单账户分红汇总
type AccountDividend struct {
	Account        string  `json:"account"`
	Holdings       int     `json:"holdings"`        // 持仓只数
	Shares         int     `json:"shares"`          // 总股数
	MarketValue    float64 `json:"market_value"`    // 市值
	AnnualDividend float64 `json:"annual_dividend"` // 年化分红金额
}

// YearDividend 单年分红汇总
type YearDividend struct {
	Year          int     `json:"year"`
	TotalDividend float64 `json:"total_dividend"` // 该年分红总额（各持仓当年每股分红×股数）
	Holdings      int     `json:"holdings"`       // 参与该年分红的持仓只数
}
