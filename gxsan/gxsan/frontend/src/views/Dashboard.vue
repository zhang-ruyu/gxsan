<template>
  <div>
    <div class="page-header">
      <h2>总览 · 养老操盘驾驶舱</h2>
      <div class="header-actions">
        <span class="last-update" v-if="lastUpdate">最后更新: {{ lastUpdate }}</span>
        <span class="refresh-state" v-if="refreshState">{{ refreshState }}</span>
        <button class="btn btn-secondary btn-sm" @click="refreshData">刷新数据</button>
      </div>
    </div>

    <div v-if="loading" class="loading">
      <div class="loading-spinner"></div>
      <p>正在汇总持仓与行动提示...</p>
    </div>

    <div v-else-if="error" class="error-state">
      <p>❌ {{ error }}</p>
      <button class="btn btn-primary" @click="refreshData">重试</button>
    </div>

    <div v-else>
      <!-- 汇总统计 -->
      <div class="summary-stats">
        <div class="stat-card">
          <div class="stat-value">¥{{ formatNumber(dash.total_assets) }}</div>
          <div class="stat-label">总资产</div>
        </div>
        <div class="stat-card">
          <div class="stat-value text-orange">¥{{ formatNumber(dash.current_annual_dividend) }}</div>
          <div class="stat-label">当前年化分红</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">¥{{ formatNumber(dash.target_annual_dividend) }}</div>
          <div class="stat-label">目标年分红</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ dash.progress_pct.toFixed(1) }}%</div>
          <div class="stat-label">退休金进度</div>
        </div>
      </div>

      <!-- 退休金进度条 -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">退休金进度 · {{ dash.lifecycle_stage }}</span>
        </div>
        <div class="progress-wrap">
          <div class="progress-bar" :style="{ width: progressWidth + '%' }"></div>
        </div>
        <div class="progress-meta">
          <span>当前年化分红 ¥{{ formatNumber(dash.current_annual_dividend) }}</span>
          <span v-if="dash.target_annual_dividend > 0">
            目标 ¥{{ formatNumber(dash.target_annual_dividend) }}（还差 ¥{{ formatNumber(remainingTarget) }}）
          </span>
          <span v-else class="text-muted">未设目标，去「系统设置」填写目标年分红</span>
        </div>
      </div>

      <!-- 行动提示汇总 -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">行动提示（共 {{ dash.holdings_count }} 只持仓）</span>
        </div>
        <div class="action-summary">
          <div class="action-chip">
            <SignalBadge signal="BUY" />
            <span class="chip-count">{{ dash.buy_count }}</span>
          </div>
          <div class="action-chip">
            <SignalBadge signal="HOLD" />
            <span class="chip-count">{{ dash.hold_count }}</span>
          </div>
          <div class="action-chip">
            <SignalBadge signal="WATCH" />
            <span class="chip-count">{{ dash.watch_count }}</span>
          </div>
          <div class="action-chip">
            <SignalBadge signal="SELL" />
            <span class="chip-count">{{ dash.sell_count }}</span>
          </div>
        </div>
      </div>

      <!-- 行动提示明细 -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">下一步动作（按优先级）</span>
        </div>
        <table class="table">
          <thead>
            <tr>
              <th>代码</th>
              <th>名称</th>
              <th>行动</th>
              <th>建议金额/数量</th>
              <th>理由</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(ad, idx) in dash.advice" :key="idx">
              <td class="font-bold">{{ ad.code }}</td>
              <td>{{ ad.name }}</td>
              <td><SignalBadge :signal="ad.action" /></td>
              <td>
                <span v-if="ad.action === 'BUY'" class="text-orange">投入 ¥{{ formatNumber(ad.suggested_buy_amount) }}</span>
                <span v-else-if="ad.action === 'SELL'" class="text-red">卖出 {{ ad.suggested_sell_shares }} 股</span>
                <span v-else>—</span>
                <div v-if="ad.constraint" class="text-muted small">{{ ad.constraint }}</div>
              </td>
              <td class="reason-cell">
                {{ ad.reason }}
                <span v-if="ad.cost_correctable" class="cost-tip">⚠️ 成本可能四舍五入失真，可在持仓页一键修正</span>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-if="dash.advice.length === 0" class="empty-state">
          <p>暂无持仓，先去「持仓管理」添加</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { GetDashboard } from '../../wailsjs/go/main/App'
import { parseJSON } from '../utils/api'
import { isPageVisible, isMarketOpen, refreshReason } from '../utils/market'
import SignalBadge from '../components/SignalBadge.vue'

export default {
  name: 'Dashboard',
  components: { SignalBadge },
  data() {
    return {
      loading: true,
      error: null,
      dash: {
        total_assets: 0,
        total_market_value: 0,
        available_fund: 0,
        current_annual_dividend: 0,
        target_annual_dividend: 0,
        progress_pct: 0,
        lifecycle_stage: '',
        holdings_count: 0,
        advice: [],
        buy_count: 0,
        sell_count: 0,
        hold_count: 0,
        watch_count: 0
      },
      lastUpdate: '',
      refreshState: '',
      refreshTimer: null
    }
  },
  computed: {
    progressWidth() {
      return Math.min(100, this.dash.progress_pct)
    },
    remainingTarget() {
      return Math.max(0, this.dash.target_annual_dividend - this.dash.current_annual_dividend)
    }
  },
  methods: {
    formatNumber(num) {
      return Number(num || 0).toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
    },
    async refreshData() {
      this.error = null
      try {
        this.refreshState = refreshReason()
        const raw = await GetDashboard()
        this.dash = parseJSON(raw)
        this.lastUpdate = new Date().toLocaleTimeString('zh-CN')
      } catch (error) {
        console.error('获取总览失败:', error)
        this.error = error.message || '获取总览失败'
      }
      this.loading = false
    },
    startAutoRefresh() {
      if (this.refreshTimer) clearInterval(this.refreshTimer)
      this.refreshTimer = setInterval(() => {
        this.refreshState = refreshReason()
        if (!isPageVisible() || !isMarketOpen()) {
          console.log('自动刷新总览跳过:', this.refreshState)
          return
        }
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

.text-orange { color: #e8590c; }
.text-red { color: var(--danger-color); }
.text-muted { color: var(--text-muted); }
.small { font-size: 12px; }

.progress-wrap {
  margin: 16px 0 8px;
  height: 18px;
  background: #eef2f7;
  border-radius: 9px;
  overflow: hidden;
}

.progress-bar {
  height: 100%;
  background: linear-gradient(90deg, #4f46e5, #22c55e);
  border-radius: 9px;
  transition: width 0.4s ease;
}

.progress-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  font-size: 13px;
  color: var(--text-muted);
}

.action-summary {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
  padding: 8px 0;
}

.action-chip {
  display: flex;
  align-items: center;
  gap: 8px;
}

.chip-count {
  font-size: 18px;
  font-weight: 700;
}

.reason-cell {
  max-width: 360px;
  line-height: 1.5;
}

.cost-tip {
  display: block;
  margin-top: 4px;
  color: #e8590c;
  font-size: 12px;
}
</style>
