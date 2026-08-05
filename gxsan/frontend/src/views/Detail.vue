<template>
  <div>
    <div class="page-header">
      <div class="header-left">
        <button class="btn btn-secondary btn-sm" @click="goBack">← 返回</button>
        <h2>{{ stockName }} ({{ stockCode }})</h2>
      </div>
      <div class="header-actions">
        <span class="last-update" v-if="lastUpdate">最后更新: {{ lastUpdate }}</span>
        <span class="refresh-state" v-if="refreshState">{{ refreshState }}</span>
        <button class="btn btn-secondary btn-sm" @click="refreshData(true)">
          刷新数据
        </button>
      </div>
    </div>

    <div v-if="loading" class="loading">
      <div class="loading-spinner"></div>
      <p>正在获取股票数据并分析...</p>
    </div>

    <div v-else-if="error" class="error-state">
      <p>❌ {{ error }}</p>
      <button class="btn btn-primary" @click="refreshData">重试</button>
    </div>

    <div v-else>
      <!-- 核心指标 -->
      <div class="metric-cards">
        <div class="metric-card">
          <div class="metric-value"><SignalBadge :signal="report.data.signal" /></div>
          <div class="metric-label">当前信号</div>
        </div>
        <div class="metric-card">
          <div class="metric-value">¥{{ report.data.current_price.toFixed(2) }}</div>
          <div class="metric-label">当前股价</div>
        </div>
        <div class="metric-card">
          <div class="metric-value">{{ report.data.latest_dividend_yield.toFixed(2) }}%</div>
          <div class="metric-label">当前股息率</div>
        </div>
        <div class="metric-card">
          <div class="metric-value">{{ report.data.avg_dividend_yield.toFixed(2) }}%</div>
          <div class="metric-label">平均股息率</div>
        </div>
        <div class="metric-card" v-if="report.data.yield_on_cost > 0">
          <div class="metric-value text-orange">{{ report.data.yield_on_cost.toFixed(2) }}%</div>
          <div class="metric-label">成本股息率</div>
        </div>
      </div>

      <!-- 估值分析 -->
      <div class="card">
        <div class="card-header">
          <span class="card-title"> </span>
        </div>
        <div class="valuation-grid">
          <div class="valuation-item">
            <div class="valuation-label">当前股息率</div>
            <div class="valuation-value" :class="currentSignalClass">
              {{ report.data.latest_dividend_yield.toFixed(2) }}%
            </div>
          </div>
          <div class="valuation-item">
            <div class="valuation-label">平均股息率</div>
            <div class="valuation-value">{{ report.data.avg_dividend_yield.toFixed(2) }}%</div>
          </div>
          <div class="valuation-item">
            <div class="valuation-label">低估阈值</div>
            <div class="valuation-value text-green">
              ≤{{ (report.data.avg_dividend_yield * report.data.cheap_discount).toFixed(2) }}%
            </div>
          </div>
          <div class="valuation-item">
            <div class="valuation-label">高估阈值</div>
            <div class="valuation-value text-red">
              ≥{{ (report.data.avg_dividend_yield * report.data.expensive_premium).toFixed(2) }}%
            </div>
          </div>
        </div>
      </div>

      <!-- 格栅策略 -->
      <div class="card">
        <div class="card-header">
          <span class="card-title"> ️ 格栅策略</span>
        </div>
        <div v-if="report.data.grid_strategy">
          <div class="grid-section">
            <h4 class="grid-title text-red">买入格栅</h4>
            <div class="grid-table">
              <div class="grid-row grid-header">
                <div class="grid-cell">股息率</div>
                <div class="grid-cell">建议金额</div>
                <div class="grid-cell">建议买入</div>
              </div>
              <div v-for="(grid, index) in report.data.grid_strategy.buy_grids" 
                   :key="'buy-' + index"
                   class="grid-row"
                   :class="{ 'active': isBuyActive(grid.yield) }">
                <div class="grid-cell font-bold">≥{{ grid.yield.toFixed(2) }}%</div>
                <div class="grid-cell">¥{{ formatNumber(grid.amount) }}</div>
                <div class="grid-cell">
                  <button v-if="isBuyActive(grid.yield)" class="btn btn-sm btn-primary" disabled>
                    建议买入
                  </button>
                  <span v-else class="text-muted">-</span>
                </div>
              </div>
            </div>
          </div>

          <div class="grid-section">
            <h4 class="grid-title text-green">卖出格栅</h4>
            <div class="grid-table">
              <div class="grid-row grid-header">
                <div class="grid-cell">股息率</div>
                <div class="grid-cell">持有股数</div>
                <div class="grid-cell">建议卖出</div>
              </div>
              <div v-for="(grid, index) in report.data.grid_strategy.sell_grids" 
                   :key="'sell-' + index"
                   class="grid-row"
                   :class="{ 'active': isSellActive(grid.yield) }">
                <div class="grid-cell font-bold">≤{{ grid.yield.toFixed(2) }}%</div>
                <div class="grid-cell">{{ formatNumber(grid.amount) }} 股</div>
                <div class="grid-cell">
                  <button v-if="isSellActive(grid.yield)" class="btn btn-sm btn-danger" disabled>
                    建议卖出
                  </button>
                  <span v-else class="text-muted">-</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div v-else class="empty-state">
          <p>暂无格栅策略</p>
        </div>
      </div>

      <!-- 历史数据 -->
      <div class="card">
        <div class="card-header">
          <span class="card-title"> ️ 历史数据</span>
        </div>
        <table class="table">
          <thead>
            <tr>
              <th>年份</th>
              <th>每股股息</th>
              <th>股息率</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(item, index) in report.data.history" :key="index">
              <td>{{ item.year }}</td>
              <td>¥{{ item.dividend_per_share.toFixed(2) }}</td>
              <td :class="item.yield > report.data.avg_dividend_yield ? 'text-green' : 'text-red'">
                {{ item.yield.toFixed(2) }}%
              </td>
            </tr>
          </tbody>
        </table>
        <div v-if="report.data.history.length === 0" class="empty-state">
          <p>暂无历史数据</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { GetStockDetail, RefreshStock } from '../../wailsjs/go/main/App'
import SignalBadge from '../components/SignalBadge.vue'
import { signalClass } from '../utils/signal'
import { parseJSON } from '../utils/api'
import { isPageVisible, isMarketOpen, refreshReason } from '../utils/market'

export default {
  name: 'Detail',
  components: {
    SignalBadge
  },
  data() {
    return {
      loading: true,
      error: null,
      report: {
        data: {
          stock_code: '',
          stock_name: '',
          current_price: 0,
          latest_dividend_yield: 0,
          avg_dividend_yield: 0,
          signal: '',
          valuation: '',
          grid_strategy: {
            buy_grids: [],
            sell_grids: []
          },
          history: [],
          cheap_discount: 0.8,
          expensive_premium: 1.2
        }
      },
      lastUpdate: '',
      refreshState: '',
      refreshTimer: null
    }
  },
  computed: {
    stockCode() {
      return this.$route.params.code
    },
    stockName() {
      return this.report.data.stock_name || this.stockCode
    },
    // 当前信号对应的 CSS 类名（估值区给「当前股息率」上色用）
    currentSignalClass() {
      return signalClass(this.report.data.signal)
    }
  },
  methods: {
    goBack() {
      this.$router.push('/')
    },
    formatNumber(num) {
      return num.toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
    },
    isBuyActive(yieldValue) {
      return this.report.data.latest_dividend_yield >= yieldValue
    },
    isSellActive(yieldValue) {
      return this.report.data.latest_dividend_yield <= yieldValue
    },
    async refreshData(force) {
      this.error = null
      try {
        this.refreshState = refreshReason()
        console.log('开始刷新股票详情:', this.stockCode, 'force=', force)
        // force=true 时绕过缓存强制重新抓数（同一支股票可随时多次刷新）
        const result = force ? await RefreshStock(this.stockCode) : await GetStockDetail(this.stockCode)
        console.log('获取到详情:', result)
        if (result) {
          this.report = parseJSON(result)
        } else {
          this.error = '未找到股票数据'
        }
        this.lastUpdate = new Date().toLocaleTimeString('zh-CN')
      } catch (error) {
        console.error('获取详情失败:', error)
        this.error = error.message || '获取详情失败'
      }
      this.loading = false
    },
    startAutoRefresh() {
      if (this.refreshTimer) {
        clearInterval(this.refreshTimer)
      }
      this.refreshTimer = setInterval(() => {
        // 智能刷新：页面隐藏或非交易时段暂停
        this.refreshState = refreshReason()
        if (!isPageVisible() || !isMarketOpen()) {
          console.log('自动刷新详情跳过:', this.refreshState)
          return
        }
        console.log('自动刷新详情触发')
        this.refreshData()
      }, 30000)
    },
    stopAutoRefresh() {
      if (this.refreshTimer) {
        clearInterval(this.refreshTimer)
        this.refreshTimer = null
      }
    }
  },
  mounted() {
    console.log('Detail组件挂载, 股票代码:', this.stockCode)
    this.refreshData()
    this.startAutoRefresh()
  },
  beforeUnmount() {
    this.stopAutoRefresh()
  }
}
</script>

<style scoped>
.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.last-update {
  font-size: 12px;
  color: var(--text-muted);
}

.refresh-state {
  font-size: 12px;
  color: var(--text-muted);
  padding: 2px 8px;
  border-radius: 10px;
  background: #f1f5f9;
}

.text-green { color: var(--success-color); }
.text-red { color: var(--danger-color); }
.text-orange { color: #e8590c; }
.text-muted { color: var(--text-muted); }
</style>
