package model

// 以下结构体为 GUI（Wails 前端）与 CLI 共用的视图模型（DTO）。
// 原先分散定义在 gxsan/app.go 中，现统一收敛到 model 包，避免重复定义。

// PortfolioItem 持仓项（GUI 展示用）
type PortfolioItem struct {
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	Shares       int     `json:"shares"`
	AvgCost      float64 `json:"avg_cost"`
	OriginalCost float64 `json:"original_cost"` // 原始买入成本快照（>0 表示曾修正过）
	Price        float64 `json:"price"`
	MarketValue  float64 `json:"market_value"`
	Profit       float64 `json:"profit"`
	ProfitPct    float64 `json:"profit_pct"`
	YieldOnCost  float64 `json:"yield_on_cost"` // 成本股息率 YoC
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

// DividendSummary 分红按年汇总（GUI 分红汇总页用）
// 按自然年汇总各持仓分红金额，便于评估分红现金流与养老本金积累。
type DividendSummary struct {
	ByYear              []YearDividend    `json:"by_year"`
	ByStock             []StockDividend   `json:"by_stock"`             // 按单只股票展现（最近一年分红）
	TotalAnnualDividend float64           `json:"total_annual_dividend"` // 当前持仓年化分红总额
	TotalMarketValue    float64           `json:"total_market_value"`
	TotalHoldings       int               `json:"total_holdings"`
}

// YearDividend 单年分红汇总
type YearDividend struct {
	Year          int     `json:"year"`
	TotalDividend float64 `json:"total_dividend"` // 该年分红总额（各持仓当年每股分红×股数）
	Holdings      int     `json:"holdings"`       // 参与该年分红的持仓只数
}

// StockDividendEvent 单笔分红事件（分红汇总页按股票展示用）
type StockDividendEvent struct {
	Date     string  `json:"date"`      // 除权除息日 YYYY-MM-DD
	PerShare float64 `json:"per_share"` // 每股派息（元/股）
	Per10    float64 `json:"per_10"`    // 每10股派息（元/10股）
	Per100   float64 `json:"per_100"`   // 每100股派息（元/100股）
	OverYear bool    `json:"over_year"` // 是否超过一年（有记录但最近一次早于近一年）
}

// StockDividend 单只股票的分红汇总（分红汇总页按股票展现用）
type StockDividend struct {
	Code             string               `json:"code"`
	Name             string               `json:"name"`
	Shares           int                  `json:"shares"`
	MarketValue      float64              `json:"market_value"`
	HasData          bool                 `json:"has_data"`           // 是否有分红记录
	RecentDividends  []StockDividendEvent `json:"recent_dividends"`  // 最近一年（或最近一次）分红
	OlderDividends   []StockDividendEvent `json:"older_dividends"`   // 更早的分红（折叠展示）
}

// ActionAdvice 单只持仓的行动提示（建议引擎 MVP 输出）
// 复用现有 divStrategy / gridStrategy 信号，给出 买/卖/等 行动 + 金额 + 理由。
type ActionAdvice struct {
	Code                string  `json:"code"`
	Name                string  `json:"name"`
	Shares              int     `json:"shares"`
	AvgCost             float64 `json:"avg_cost"`
	Price               float64 `json:"price"`
	YieldOnCost         float64 `json:"yield_on_cost"`  // 成本股息率
	CurrentYield        float64 `json:"current_yield"`  // 当前静态股息率
	TargetYield         float64 `json:"target_yield"`   // 目标股息率
	Action              string  `json:"action"`         // BUY / HOLD / SELL / WATCH
	SuggestedBuyAmount  float64 `json:"suggested_buy_amount"`  // ¥（Action=BUY 时有效）
	SuggestedBuyShares  int     `json:"suggested_buy_shares"`  // 估算可买股数（A股按100股向下取整）
	SuggestedSellShares int     `json:"suggested_sell_shares"` // 股（Action=SELL 时有效）
	SuggestedSellAmount float64 `json:"suggested_sell_amount"` // 估算卖出可得金额
	Reason              string  `json:"reason"`
	Constraint          string  `json:"constraint"` // 约束提示（资金不足/已达上限/未设网格等）
	Priority            int     `json:"priority"`   // 数字越小越优先执行
	CheapPrice          float64 `json:"cheap_price"`
	FairPrice           float64 `json:"fair_price"`
	CostCorrectable     bool    `json:"cost_correctable"` // 成本可能因四舍五入失真（尚无真实投入总额记录），提示可做一键修正
}
