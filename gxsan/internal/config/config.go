package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/gxsan/internal/model"
	"gopkg.in/yaml.v3"
)

const (
	ConfigDir  = "gxsan"
	ConfigFile = "config.yaml"
	DataDir    = "data"
)

// Manager 配置管理器
type Manager struct {
	Config     *model.Config
	ConfigPath string
	DataPath   string
}

// NewManager 创建配置管理器
func NewManager() *Manager {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ConfigDir)
	dataDir := filepath.Join(configDir, DataDir)

	os.MkdirAll(configDir, 0755)
	os.MkdirAll(dataDir, 0755)

	return &Manager{
		ConfigPath: filepath.Join(configDir, ConfigFile),
		DataPath:   dataDir,
	}
}

// Load 加载配置
func (m *Manager) Load() error {
	if _, err := os.Stat(m.ConfigPath); os.IsNotExist(err) {
		m.Config = model.DefaultConfig()
		return m.Save()
	}

	data, err := os.ReadFile(m.ConfigPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	m.Config = model.DefaultConfig()
	if err := yaml.Unmarshal(data, m.Config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	return nil
}

// Save 保存配置
func (m *Manager) Save() error {
	data, err := yaml.Marshal(m.Config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(m.ConfigPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// GetStockConfig 获取股票配置
func (m *Manager) GetStockConfig(code string) *model.StockConfig {
	for i := range m.Config.Watchlist {
		if m.Config.Watchlist[i].Code == code {
			return &m.Config.Watchlist[i]
		}
	}
	return nil
}

// AddStock 添加股票
func (m *Manager) AddStock(code, name string, targetYield float64) error {
	for _, s := range m.Config.Watchlist {
		if s.Code == code {
			return fmt.Errorf("股票 %s 已存在", code)
		}
	}

	stock := model.StockConfig{
		Code:        code,
		Name:        name,
		TargetYield: targetYield,
		GridStrategy: model.GridStrategy{
			BuyGrids: [5]model.GridLevel{
				{Yield: targetYield + 0.5, Amount: 5000},
				{Yield: targetYield + 1.0, Amount: 8000},
				{Yield: targetYield + 1.5, Amount: 10000},
				{Yield: targetYield + 2.0, Amount: 15000},
				{Yield: targetYield + 2.5, Amount: 20000},
			},
			SellGrids: [5]model.GridLevel{
				{Yield: targetYield - 0.5, Amount: 20},
				{Yield: targetYield - 1.0, Amount: 30},
				{Yield: targetYield - 1.5, Amount: 50},
				{Yield: targetYield - 2.0, Amount: 80},
				{Yield: targetYield - 2.5, Amount: 100},
			},
		},
	}

	m.Config.Watchlist = append(m.Config.Watchlist, stock)
	return m.Save()
}

// RemoveStock 删除股票
func (m *Manager) RemoveStock(code string) error {
	for i, s := range m.Config.Watchlist {
		if s.Code == code {
			m.Config.Watchlist = append(m.Config.Watchlist[:i], m.Config.Watchlist[i+1:]...)
			return m.Save()
		}
	}
	return fmt.Errorf("股票 %s 不存在", code)
}

// UpdateStock 更新监控股票的名称与目标股息率
func (m *Manager) UpdateStock(code, name string, targetYield float64) error {
	for i := range m.Config.Watchlist {
		if m.Config.Watchlist[i].Code == code {
			if name != "" {
				m.Config.Watchlist[i].Name = name
			}
			m.Config.Watchlist[i].TargetYield = targetYield
			return m.Save()
		}
	}
	return fmt.Errorf("股票 %s 不存在", code)
}

// AddHolding 添加持仓
func (m *Manager) AddHolding(code, name string, shares int, avgCost float64) error {
	for i, h := range m.Config.Portfolio {
		if h.Code == code {
			m.Config.Portfolio[i].Shares = shares
			m.Config.Portfolio[i].AvgCost = avgCost
			return m.Save()
		}
	}

	holding := model.Holding{
		Code:    code,
		Name:    name,
		Shares:  shares,
		AvgCost: avgCost,
	}

	m.Config.Portfolio = append(m.Config.Portfolio, holding)
	return m.Save()
}

// UpdateHolding 更新持仓
func (m *Manager) UpdateHolding(code string, shares int) error {
	for i, h := range m.Config.Portfolio {
		if h.Code == code {
			m.Config.Portfolio[i].Shares = shares
			return m.Save()
		}
	}
	return fmt.Errorf("持仓 %s 不存在", code)
}

// RemoveHolding 删除持仓
func (m *Manager) RemoveHolding(code string) error {
	for i, h := range m.Config.Portfolio {
		if h.Code == code {
			m.Config.Portfolio = append(m.Config.Portfolio[:i], m.Config.Portfolio[i+1:]...)
			return m.Save()
		}
	}
	return fmt.Errorf("持仓 %s 不存在", code)
}

// SetFund 设置资金
func (m *Manager) SetFund(available float64) error {
	m.Config.Fund.AvailableFund = available
	return m.Save()
}

// SetMaxPosition 设置单只股票最大持仓占比
func (m *Manager) SetMaxPosition(pct float64) error {
	m.Config.Fund.MaxPositionPct = pct
	return m.Save()
}

// SetLifecycleStage 设置生命周期阶段（1启动 2滚雪球 3自由 4收获）
func (m *Manager) SetLifecycleStage(stage int) error {
	if stage < 1 || stage > 4 {
		return fmt.Errorf("生命周期阶段须在 1-4 之间")
	}
	m.Config.LifecycleStage = stage
	return m.Save()
}

// CorrectHoldingCost 一键修正持仓成本：以真实投入总额推导精确每股成本
// 解决 A股「总价÷股数」产生多位小数、手动录入每股成本被四舍五入失真的问题。
func (m *Manager) CorrectHoldingCost(code string, totalCost float64) error {
	for i, h := range m.Config.Portfolio {
		if h.Code == code {
			if h.Shares <= 0 {
				return fmt.Errorf("持仓 %s 股数为 0，无法修正成本", code)
			}
			if totalCost < 0 {
				return fmt.Errorf("投入总额不能为负")
			}
			m.Config.Portfolio[i].TotalCost = totalCost
			m.Config.Portfolio[i].AvgCost = totalCost / float64(h.Shares) // 精确推导，保留 float64 全精度
			return m.Save()
		}
	}
	return fmt.Errorf("持仓 %s 不存在", code)
}

// SetGrid 设置网格
func (m *Manager) SetGrid(code string, isBuy bool, level int, yield float64, amount float64) error {
	stock := m.GetStockConfig(code)
	if stock == nil {
		return fmt.Errorf("股票 %s 不存在", code)
	}

	if level < 1 || level > 5 {
		return fmt.Errorf("档位必须在 1-5 之间")
	}

	if isBuy {
		stock.GridStrategy.BuyGrids[level-1] = model.GridLevel{
			Yield:  yield,
			Amount: amount,
		}
	} else {
		stock.GridStrategy.SellGrids[level-1] = model.GridLevel{
			Yield:  yield,
			Amount: amount,
		}
	}

	return m.Save()
}
