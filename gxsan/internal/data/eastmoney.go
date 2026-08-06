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
	apiURL := fmt.Sprintf("%s?secid=%s&fields=f43,f44,f45,f46,f47,f48,f50,f51,f52,f55,f57,f58,f116,f117,f162,f167,f170", eastMoneyQuoteURL, secid)

	// 行情接口偶发限流/校验失败，最多重试 3 次（指数退避），大幅提升成功率。
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 300 * time.Millisecond)
		}
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %w", err)
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		req.Header.Set("Referer", "https://quote.eastmoney.com/")
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
		req.Header.Set("Connection", "keep-alive")

		resp, err := f.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("请求失败(第%d次): %w", attempt+1, err)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("读取响应失败(第%d次): %w", attempt+1, err)
			continue
		}

		// 容错解析：先整体解析；若某个字段类型异常导致整体失败，再用 RawMessage 单独取关键字段。
		var parsed struct {
			Data struct {
				F43  float64 `json:"f43"`
				F57  string  `json:"f57"`
				F58  string  `json:"f58"`
				F162 float64 `json:"f162"`
				F167 float64 `json:"f167"`
				F170 float64 `json:"f170"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &parsed); err != nil {
			var raw struct {
				Data json.RawMessage `json:"data"`
			}
			if jerr := json.Unmarshal(body, &raw); jerr != nil {
				bs := string(body)
				if len(bs) > 200 {
					bs = bs[:200]
				}
				lastErr = fmt.Errorf("解析数据失败(第%d次): %w | body=%s", attempt+1, err, bs)
				continue
			}
			var d struct {
				F43 float64 `json:"f43"`
				F57 string  `json:"f57"`
				F58 string  `json:"f58"`
			}
			if jerr := json.Unmarshal(raw.Data, &d); jerr != nil {
				lastErr = fmt.Errorf("解析行情字段失败(第%d次): %w", attempt+1, jerr)
				continue
			}
			parsed.Data.F43, parsed.Data.F57, parsed.Data.F58 = d.F43, d.F57, d.F58
		}

		// 接口偶发返回空 data（限流/校验失败/网络抖动）时 F57 为空且 F43 为 0；
		// 若不当作失败处理，会被当成"合法 0 价"写入缓存，导致持仓页股价归零。
		if parsed.Data.F57 == "" || parsed.Data.F43 == 0 {
			lastErr = fmt.Errorf("行情返回空数据(第%d次, 代码=%q F43=%.2f): %s", attempt+1, parsed.Data.F57, parsed.Data.F43, code)
			continue
		}

		return &model.Stock{
			Code:   parsed.Data.F57,
			Name:   parsed.Data.F58,
			Price:  parsed.Data.F43 / 100, // 东方财富价格单位是分
			Change: parsed.Data.F170 / 100,
			PE:     parsed.Data.F162 / 100,
			PB:     parsed.Data.F167 / 100,
		}, nil
	}
	return nil, lastErr
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
