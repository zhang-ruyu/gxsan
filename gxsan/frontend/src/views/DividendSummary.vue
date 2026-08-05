<template>
  <div>
    <div class="page-header">
      <h2>分红汇总</h2>
      <div class="header-actions">
        <span class="last-update" v-if="lastUpdate">最后更新: {{ lastUpdate }}</span>
        <span class="refresh-state" v-if="refreshState">{{ refreshState }}</span>
        <button class="btn btn-secondary btn-sm" @click="refreshData">刷新数据</button>
      </div>
    </div>

    <div v-if="loading" class="loading">
      <div class="loading-spinner"></div>
      <p>正在汇总分红数据...</p>
    </div>

    <div v-else-if="error" class="error-state">
      <p>❌ {{ error }}</p>
      <button class="btn btn-primary" @click="refreshData">重试</button>
    </div>

    <div v-else>
      <!-- 汇总统计 -->
      <div class="summary-stats">
        <div class="stat-card">
          <div class="stat-value">{{ summary.total_holdings }}</div>
          <div class="stat-label">持仓只数</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">¥{{ formatNumber(summary.total_market_value) }}</div>
          <div class="stat-label">持仓市值</div>
        </div>
        <div class="stat-card">
          <div class="stat-value text-orange">¥{{ formatNumber(summary.total_annual_dividend) }}</div>
          <div class="stat-label">年化分红</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ overallYield }}%</div>
          <div class="stat-label">组合股息率</div>
        </div>
      </div>

      <!-- 按年汇总 -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">按自然年汇总（历年分红 × 当前持仓股数）</span>
        </div>
        <table class="table">
          <thead>
            <tr>
              <th>年份</th>
              <th>参与持仓</th>
              <th>分红总额</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(yr, index) in summary.by_year" :key="index">
              <td class="font-bold">{{ yr.year }}</td>
              <td>{{ yr.holdings }}</td>
              <td class="text-orange">¥{{ formatNumber(yr.total_dividend) }}</td>
            </tr>
          </tbody>
        </table>
        <div v-if="summary.by_year.length === 0" class="empty-state">
          <p>暂无分红记录</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { GetDividendSummary } from '../../wailsjs/go/main/App'
import { parseJSON } from '../utils/api'
import { isPageVisible, isMarketOpen, refreshReason } from '../utils/market'

export default {
  name: 'DividendSummary',
  data() {
    return {
      loading: true,
      error: null,
      summary: {
        by_year: [],
        total_annual_dividend: 0,
        total_market_value: 0,
        total_holdings: 0
      },
      lastUpdate: '',
      refreshState: '',
      refreshTimer: null
    }
  },
  computed: {
    overallYield() {
      const mv = this.summary.total_market_value
      if (mv > 0) {
        return (this.summary.total_annual_dividend / mv * 100).toFixed(2)
      }
      return '0.00'
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
        const raw = await GetDividendSummary()
        this.summary = parseJSON(raw)
        this.lastUpdate = new Date().toLocaleTimeString('zh-CN')
      } catch (error) {
        console.error('获取分红汇总失败:', error)
        this.error = error.message || '获取分红汇总失败'
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
          console.log('自动刷新分红汇总跳过:', this.refreshState)
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

.text-green { color: var(--success-color); }
.text-orange { color: #e8590c; }
</style>
