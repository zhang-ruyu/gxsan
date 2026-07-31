# 股息三 (GxSan)

A股股息分析与投资决策辅助工具

## 功能特点

- **股息率分析**: 基于股息率判断股票价值
- **网格投资策略**: 买入/卖出各5档网格
- **投资组合管理**: 手动输入持仓，跟踪收益
- **资金监控**: 根据可用资金推荐操作
- **股利日历**: 显示即将除息的股票

## 安装

### 方式一：从源码编译

```bash
# 需要先安装 Go 1.21+
# 下载地址: https://golang.org/dl/

# 克隆项目
git clone https://github.com/user/gxsan.git
cd gxsan

# 编译
go build -o gxsan.exe ./cmd

# 添加到 PATH 环境变量
```

### 方式二：直接下载

从 Releases 页面下载预编译的可执行文件。

## 快速开始

```bash
# 1. 添加监控股票
gxsan add 601398 工商银行 --target-yield 5.0
gxsan add 601288 农业银行 --target-yield 5.0
gxsan add 600519 贵州茅台 --target-yield 2.0

# 2. 设置可用资金
gxsan fund set 50000

# 3. 添加持仓
gxsan portfolio add 601398 --shares 1000 --cost 5.00
gxsan portfolio add 601288 --shares 2000 --cost 3.60

# 4. 运行分析
gxsan analyze

# 5. 查看单只股票详情
gxsan detail 601398

# 6. 查看股利日历
gxsan calendar
```

## 命令列表

### 股票管理

| 命令 | 说明 |
|------|------|
| `gxsan add <代码> <名称>` | 添加股票到监控列表 |
| `gxsan list` | 查看所有监控股票 |
| `gxsan remove <代码>` | 删除股票 |
| `gxsan search <关键词>` | 搜索股票 |

### 网格策略

| 命令 | 说明 |
|------|------|
| `gxsan grid set <代码> --buy <档位> <股息率> <金额>` | 设置买入网格 |
| `gxsan grid set <代码> --sell <档位> <股息率> <比例>` | 设置卖出网格 |
| `gxsan grid show <代码>` | 查看网格策略 |

### 投资组合

| 命令 | 说明 |
|------|------|
| `gxsan portfolio add <代码> --shares <股数> --cost <成本价>` | 添加持仓 |
| `gxsan portfolio update <代码> --shares <股数>` | 更新持仓 |
| `gxsan portfolio list` | 查看持仓列表 |
| `gxsan portfolio remove <代码>` | 删除持仓 |

### 资金管理

| 命令 | 说明 |
|------|------|
| `gxsan fund set <金额>` | 设置可用资金 |
| `gxsan fund recommend` | 查看操作推荐 |

### 分析报告

| 命令 | 说明 |
|------|------|
| `gxsan analyze` | 运行分析并生成报告 |
| `gxsan detail <代码>` | 查看单只股票详情 |
| `gxsan calendar` | 查看股利日历 |

### 配置管理

| 命令 | 说明 |
|------|------|
| `gxsan config show` | 查看当前配置 |
| `gxsan config set <键> <值>` | 修改配置 |

## 配置说明

配置文件位于 `~/.gxsan/config.yaml`

### 配置项

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `default_target_yield` | 默认目标股息率 | 4.0% |
| `min_dividend_years` | 最少分红年数 | 3年 |
| `cheap_discount` | 便宜价折扣 | 80% |
| `expensive_premium` | 昂贵价溢价 | 120% |
| `fund.available_fund` | 可用资金 | 0 |
| `fund.max_position_pct` | 单只股票最大持仓占比 | 30% |

## 网格策略说明

### 买入网格

当股息率达到某档阈值时，建议投入对应金额：

| 档位 | 股息率阈值 | 建议投入 |
|------|------------|----------|
| 1 | >= 5.5% | ¥5,000 |
| 2 | >= 6.0% | ¥8,000 |
| 3 | >= 6.5% | ¥10,000 |
| 4 | >= 7.0% | ¥15,000 |
| 5 | >= 7.5% | ¥20,000 |

### 卖出网格

当股息率低于某档阈值时，建议卖出对应比例：

| 档位 | 股息率阈值 | 卖出比例 |
|------|------------|----------|
| 1 | <= 4.5% | 20% |
| 2 | <= 4.0% | 30% |
| 3 | <= 3.5% | 50% |
| 4 | <= 3.0% | 80% |
| 5 | <= 2.5% | 100% |

## 数据来源

- 实时行情: 东方财富API
- 分红数据: 东方财富数据中心

## 免责声明

本工具仅供参考，不构成任何投资建议。股市有风险，投资需谨慎。

## 许可证

MIT License
