package model

// Signal 交易信号
type Signal struct {
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	Action         string  `json:"action"`
	Price          float64 `json:"price"`
	CurrentYield   float64 `json:"current_yield"`
	TargetYield    float64 `json:"target_yield"`
	FairPrice      float64 `json:"fair_price"`
	CheapPrice     float64 `json:"cheap_price"`
	ExpensivePrice float64 `json:"expensive_price"`
	Score          int     `json:"score"`
	Reason         string  `json:"reason"`
	Timestamp      string  `json:"timestamp"`
}

// GridSignal 网格信号
type GridSignal struct {
	Code         string
	Name         string
	CurrentYield float64
	BuyLevel     int
	BuyAmount    float64
	SellLevel    int
	SellPercent  float64
	Action       string
	Reason       string
}
