<template>
  <div>
    <div class="page-header">
      <h2>分析首页</h2>
      <div class="header-actions">
        <span class="last-update" v-if="lastUpdate">{{ lastUpdate }}</span>
        <span class="refresh-state" v-if="refreshState">{{ refreshState }}</span>
        <button class="btn btn-secondary btn-sm" @click="refreshData">刷新数据</button>
      </div>
    </div>

    <div v-if="loading" class="loading">
      <div class="loading-spinner"></div>
      <p>正在汇总数据...</p>
    </div>

    <div v-else>
      <!-- 1. 资金快照 -->
      <div class="summary-stats">
        <div class="stat-card highlight">
          <div class="stat-value">¥{{ formatNumber(dash.available_fund) }}</div>
          <div class="stat-label">可用资金</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">¥{{ formatNumber(dash.total_market_value) }}</div>
          <div class="stat-label">持仓市值</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">¥{{ formatNumber(dash.total_assets) }}</div>
          <div class="stat-label">总资产</div>
        </div>
        <div class="stat-card">
          <div class="stat-value text-orange">¥{{ formatNumber(dash.current_annual_dividend) }}</div>
          <div class="stat-label">年化分红</div>
        </div>
      </div>

      <!-- 2. 持仓行动提示 -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">
            持仓行动提示
            <span v-if="dash.holdings_count > 0" class="text-muted small">
              （{{ dash.holdings_count }} 只 · 建议买入 {{ dash.buy_count }} · 建议卖出 {{ dash.sell_count }} · 持有 {{ dash.hold_count }}）
            </span>
          </span>
          <router-link to="/portfolio" class="link-btn">管理持仓 →</router-link>
        </div>

        <div v-if="dash.advice && dash.advice.length > 0" class="action-list">
          <div
            v-for="ad in sortedAdvice"
            :key="ad.code"
            class="action-row"
            :class="'action-' + ad.action.toLowerCase()"
            @click="goToDetail(ad.code)"
          >
            <div class="action-left">
              <SignalBadge :signal="ad.action" />
              <div class="action-stock">
                <span class="stock-name">{{ ad.name }}</span>
                <span class="stock-code">{{ ad.code }}</span>
                <span class="stock-shares">{{ ad.shares }} 股</span>
              </div>
            </div>
            <div class="action-mid">
              <div class="yield-info">
                <span :class="ad.current_yield >= ad.target_yield ? 'text-green' : ''">
                  {{ ad.current_yield.toFixed(2) }}%
                </span>
                <span class="text-muted small">目标 {{ ad.target_yield.toFixed(2) }}%</span>
              </div>
            </div>
            <div class="action-right">
              <div v-if="ad.action === 'BUY'" class="action-amount text-orange">
                投入 ¥{{ formatNumber(ad.suggested_buy_amount) }}
                <span v-if="ad.suggested_buy_shares > 0" class="text-muted small">（约 {{ ad.suggested_buy_shares }} 股）</span>
              </div>
              <div v-else-if="ad.action === 'SELL'" class="action-amount text-red">
                卖出 {{ ad.suggested_sell_shares }} 股
                <span class="text-muted small">（约 ¥{{ formatNumber(ad.suggested_sell_amount) }}）</span>
              </div>
              <div v-else class="action-amount text-muted">—</div>
              <div class="action-reason text-muted small">{{ ad.reason }}</div>
            </div>
          </div>
        </div>

        <div v-else class="empty-state">
          <p>暂无持仓</p>
          <p class="text-muted">去「持仓管理」添加持仓后，这里会显示每只股票的买卖建议</p>
          <router-link to="/portfolio" class="btn btn-primary btn-sm">去添加持仓</router-link>
        </div>
      </div>

      <!-- 3. 跟踪推荐概要 -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">
            跟踪推荐
            <span class="text-muted small">（{{ trackingTotal }} 只标的 · {{ trackingCategories.length }} 大分类）</span>
          </span>
          <router-link to="/tracking" class="link-btn">查看全部 →</router-link>
        </div>

        <div class="tracking-summary">
          <div
            v-for="cat in trackingCategories"
            :key="cat.name"
            class="tracking-cat-tag"
            @click="goToTracking"
          >
            <span class="cat-name">{{ cat.name }}</span>
            <span class="cat-count">{{ cat.stocks.length }}</span>
          </div>
        </div>

        <div v-if="topYieldStocks.length > 0" class="top-yield-list">
          <div class="top-yield-title">高息标的</div>
          <div class="top-yield-grid">
            <div
              v-for="s in topYieldStocks"
              :key="s.code"
              class="top-yield-card"
              @click="goToDetail(s.code)"
            >
              <div class="ty-name">{{ s.name }}</div>
              <div class="ty-code">{{ s.code }}</div>
              <div class="ty-yield" :class="s.current_yield >= 5 ? 'text-red' : ''">
                {{ s.current_yield > 0 ? s.current_yield.toFixed(2) + '%' : '—' }}
              </div>
              <div class="ty-type" :class="s.asset_type === '弱周期' ? 'tag-weak' : 'tag-strong'">
                {{ s.asset_type }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { GetDashboard, GetTrackingStocks } from '../../wailsjs/go/main/App'
import { parseJSON } from '../utils/api'
import { isPageVisible, isMarketOpen, refreshReason } from '../utils/market'
import SignalBadge from '../components/SignalBadge.vue'

export default {
  name: 'Home',
  components: { SignalBadge },
  data() {
    return {
      loading: true,
      dash: {
        total_assets: 0,
        total_market_value: 0,
        available_fund: 0,
        current_annual_dividend: 0,
        holdings_count: 0,
        advice: [],
        buy_count: 0,
        sell_count: 0,
        hold_count: 0,
        watch_count: 0
      },
      trackingCategories: [],
      lastUpdate: '',
      refreshState: '',
      refreshTimer: null
    }
  },
  computed: {
    sortedAdvice() {
      if (!this.dash.advice) return []
      const priority = { BUY: 0, SELL: 1, WATCH: 2, HOLD: 3 }
      return [...this.dash.advice].sort((a, b) => {
        const pa = priority[a.action] ?? 9
        const pb = priority[b.action] ?? 9
        if (pa !== pb) return pa - pb
        return (b.priority || 0) - (a.priority || 0)
      })
    },
    trackingTotal() {
      return this.trackingCategories.reduce((sum, c) => sum + c.stocks.length, 0)
    },
    topYieldStocks() {
      const all = []
      for (const cat of this.trackingCategories) {
        for (const s of cat.stocks) {
          all.push({ ...s, category: cat.name })
        }
      }
      return all
        .filter(s => s.current_yield > 0)
        .sort((a, b) => b.current_yield - a.current_yield)
        .slice(0, 6)
    }
  },
  methods: {
    formatNumber(num) {
      return Number(num || 0).toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
    },
    goToDetail(code) {
      this.$router.push(`/detail/${code}`)
    },
    goToTracking() {
      this.$router.push('/tracking')
    },
    async refreshData() {
      try {
        this.refreshState = refreshReason()
        const [dashStr, trackStr] = await Promise.all([
          GetDashboard(),
          GetTrackingStocks()
        ])
        this.dash = parseJSON(dashStr)
        this.trackingCategories = parseJSON(trackStr) || []
        this.lastUpdate = new Date().toLocaleTimeString('zh-CN')
      } catch (error) {
        console.error('获取数据失败:', error)
      }
      this.loading = false
    },
    startAutoRefresh() {
      if (this.refreshTimer) clearInterval(this.refreshTimer)
      this.refreshTimer = setInterval(() => {
        this.refreshState = refreshReason()
        if (!isPageVisible() || !isMarketOpen()) return
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
    this.refreshData()
    this.startAutoRefresh()
  },
  beforeUnmount() {
    this.stopAutoRefresh()
  }
}
</script>

<style scoped>
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

/* 资金快照 */
.stat-card.highlight {
  background: linear-gradient(135deg, #4f46e5, #6366f1);
  color: #fff;
  border: none;
}
.stat-card.highlight .stat-value { color: #fff; }
.stat-card.highlight .stat-label { color: rgba(255, 255, 255, 0.85); }

.text-green { color: var(--success-color); }
.text-red { color: var(--danger-color); }
.text-orange { color: #e8590c; }
.text-muted { color: var(--text-muted); }
.small { font-size: 12px; }

.link-btn {
  font-size: 13px;
  color: #4f46e5;
  text-decoration: none;
  white-space: nowrap;
}
.link-btn:hover { text-decoration: underline; }

/* 持仓行动列表 */
.action-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.action-row {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 8px;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
  transition: background 0.15s ease;
}
.action-row:hover { background: #f8f9fa; }
.action-row:last-child { border-bottom: none; }

.action-row.action-buy { border-left: 3px solid #22c55e; }
.action-row.action-sell { border-left: 3px solid #ef4444; }
.action-row.action-watch { border-left: 3px solid #f59e0b; }
.action-row.action-hold { border-left: 3px solid #94a3b8; }

.action-left {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 200px;
}

.action-stock {
  display: flex;
  flex-direction: column;
}
.stock-name { font-weight: 600; font-size: 14px; }
.stock-code { font-size: 12px; color: var(--text-muted); }
.stock-shares { font-size: 12px; color: var(--text-muted); }

.action-mid {
  min-width: 120px;
}
.yield-info {
  display: flex;
  flex-direction: column;
  font-size: 14px;
}

.action-right {
  flex: 1;
  text-align: right;
}
.action-amount { font-weight: 600; font-size: 14px; }
.action-reason {
  margin-top: 2px;
  line-height: 1.4;
  max-width: 360px;
  margin-left: auto;
}

/* 跟踪推荐概要 */
.tracking-summary {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 20px;
}

.tracking-cat-tag {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border-radius: 20px;
  background: #f1f5f9;
  font-size: 13px;
  cursor: pointer;
  transition: background 0.15s ease;
}
.tracking-cat-tag:hover { background: #e2e8f0; }
.cat-name { color: #475569; }
.cat-count {
  background: #4f46e5;
  color: #fff;
  border-radius: 10px;
  padding: 1px 8px;
  font-size: 11px;
  font-weight: 600;
}

.top-yield-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 10px;
}

.top-yield-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(130px, 1fr));
  gap: 10px;
}

.top-yield-card {
  padding: 12px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
  text-align: center;
}
.top-yield-card:hover {
  border-color: #6366f1;
  box-shadow: 0 2px 6px rgba(99, 102, 241, 0.1);
}

.ty-name { font-weight: 600; font-size: 14px; }
.ty-code { font-size: 11px; color: var(--text-muted); margin-bottom: 6px; }
.ty-yield { font-size: 18px; font-weight: 700; color: #22c55e; }
.ty-type {
  display: inline-block;
  margin-top: 6px;
  padding: 2px 8px;
  border-radius: 8px;
  font-size: 11px;
}
.tag-weak { background: #dbeafe; color: #1d4ed8; }
.tag-strong { background: #fef3c7; color: #b45309; }
</style>
