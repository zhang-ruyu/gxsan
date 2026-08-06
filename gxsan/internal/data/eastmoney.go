package data

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/user/gxsan/internal/model"
)

const (
	eastMoneyQuoteURL = "https://push2.eastmoney.com/api/qt/stock/get"
	eastMoneyDivURL   = "https://datacenter-web.eastmoney.com/api/data/v1/get"
	eastMoneySearchURL = "https://searchapi.eastmoney.com/api/suggest/get"
)

// EastMoneyFetcher 东方财富数据获取器
type EastMoneyFetcher struct {
	client *http.Client
}

// NewEastMoneyFetcher 创建东方财富数据获取器
func NewEastMoneyFetcher() *EastMoneyFetcher {
	return &EastMoneyFetcher{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// getSecID 获取市场代码
func getSecID(code string) string {
	if strings.HasPrefix(code, "6") {
		return "1." + code
	}
	return "0." + code
}

// GetStock 获取股票实时数据
func (f *EastMoneyFetcher) GetStock(code string) (*model.Stock, error) {
	secid := getSecID(code)
	url := fmt.Sprintf("%s?secid=%s&fields=f43,f44,f45,f46,f47,f48,f50,f51,f52,f55,f57,f58,f116,f117,f162,f167,f170", eastMoneyQuoteURL, secid)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://quote.eastmoney.com/")
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var result struct {
		Data struct {
			F43  float64 `json:"f43"`  // 最新价
			F44  float64 `json:"f44"`  // 最高
			F45  float64 `json:"f45"`  // 最低
			F46  float64 `json:"f46"`  // 开盘价
			F47  float64 `json:"f47"`  // 成交量
			F48  float64 `json:"f48"`  // 成交额
			F57  string  `json:"f57"`  // 代码
			F58  string  `json:"f58"`  // 名称
			F116 float64 `json:"f116"` // 总市值
			F117 float64 `json:"f117"` // 流通市值
			F162 float64 `json:"f162"` // 市盈率
			F167 float64 `json:"f167"` // 市净率
			F170 float64 `json:"f170"` // 涨跌幅
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析数据失败: %w", err)
	}

	// 接口偶发返回空 data（限流/校验失败/网络抖动）时 F57 为空且 F43 为 0；
	// 若不当作失败处理，会被当成"合法 0 价"写入缓存，导致持仓页股价归零。
	if result.Data.F57 == "" || result.Data.F43 == 0 {
		return nil, fmt.Errorf("行情返回空数据(代码=%q F43=%.2f): %s", result.Data.F57, result.Data.F43, code)
	}

	stock := &model.Stock{
		Code:   result.Data.F57,
		Name:   result.Data.F58,
		Price:  result.Data.F43 / 100, // 东方财富价格单位是分
		Change: result.Data.F170 / 100,
		PE:     result.Data.F162 / 100,
		PB:     result.Data.F167 / 100,
	}

	return stock, nil
}

// GetDividendHistory 获取分红历史
func (f *EastMoneyFetcher) GetDividendHistory(code string) ([]model.DividendRecord, error) {
	filter := fmt.Sprintf(`(SECURITY_CODE="%s")`, code)
	encodedFilter := url.QueryEscape(filter)
	apiURL := fmt.Sprintf("%s?reportName=RPT_SHAREBONUS_DET&filter=%s&columns=ALL&pageSize=10&sortColumns=EX_DIVIDEND_DATE&sortTypes=-1", eastMoneyDivURL, encodedFilter)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://data.eastmoney.com/")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var result struct {
		Result struct {
			Data []struct {
				ExDividendDate string  `json:"EX_DIVIDEND_DATE"`
				BonusAmount    float64 `json:"BONUS_IT_RATIO"`
				CashAmount     float64 `json:"PRETAX_BONUS_RMB"`
			} `json:"data"`
		} `json:"result"`
		Success bool `json:"success"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析数据失败: %w", err)
	}

	if !result.Success {
		// 尝试备用API
		return f.getDividendHistoryBackup(code)
	}

	var records []model.DividendRecord
	for _, r := range result.Result.Data {
		if len(r.ExDividendDate) >= 10 {
			records = append(records, model.DividendRecord{
				Date:      r.ExDividendDate[:10],
				Amount:    r.CashAmount / 10, // API返回的是每股10股的派息，需要除以10
				ExDivDate: r.ExDividendDate[:10],
			})
		}
	}

	return records, nil
}

// getDividendHistoryBackup 备用分红数据获取
func (f *EastMoneyFetcher) getDividendHistoryBackup(code string) ([]model.DividendRecord, error) {
	filter := fmt.Sprintf(`(SECURITY_CODE="%s")`, code)
	encodedFilter := url.QueryEscape(filter)
	apiURL := fmt.Sprintf("%s?reportName=RPT_DMSK_FN_CASHDIVIDEND&filter=%s&columns=ALL&pageSize=10&sortColumns=EX_DIVIDEND_DATE&sortTypes=-1", eastMoneyDivURL, encodedFilter)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://data.eastmoney.com/")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var result struct {
		Result struct {
			Data []struct {
				ExDividendDate string  `json:"EX_DIVIDEND_DATE"`
				CashAmount     float64 `json:"DIVIDEND_CASH_BEFORE_TAX"`
			} `json:"data"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析数据失败: %w", err)
	}

	var records []model.DividendRecord
	for _, r := range result.Result.Data {
		if len(r.ExDividendDate) >= 10 {
			records = append(records, model.DividendRecord{
				Date:      r.ExDividendDate[:10],
				Amount:    r.CashAmount / 10, // API返回的是每股10股的派息，需要除以10
				ExDivDate: r.ExDividendDate[:10],
			})
		}
	}

	return records, nil
}

// SearchStock 搜索股票
func (f *EastMoneyFetcher) SearchStock(keyword string) ([]model.StockInfo, error) {
	url := fmt.Sprintf("%s?input=%s&type=14&token=D43BF722C8E33BDC906FB84D85E326E8&count=10", eastMoneySearchURL, keyword)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://quote.eastmoney.com/")
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var result struct {
		QuotationCodeTable struct {
			Data []struct {
				Code     string `json:"Code"`
				Name     string `json:"Name"`
				Classify string `json:"Classify"`
			} `json:"Data"`
		} `json:"QuotationCodeTable"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析数据失败: %w", err)
	}

	var stocks []model.StockInfo
	for _, s := range result.QuotationCodeTable.Data {
		// 只返回A股股票
		if s.Classify == "AStock" {
			stocks = append(stocks, model.StockInfo{
				Code: s.Code,
				Name: s.Name,
			})
		}
	}

	return stocks, nil
}

// CalculateDividendYield 计算股息率
func CalculateDividendYield(price, dividendPerShare float64) float64 {
	if price == 0 {
		return 0
	}
	return dividendPerShare / price * 100
}

// CalculateFairPrice 计算合理价
func CalculateFairPrice(dividendPerShare, targetYield float64) float64 {
	if targetYield == 0 {
		return 0
	}
	return dividendPerShare / (targetYield / 100)
}
