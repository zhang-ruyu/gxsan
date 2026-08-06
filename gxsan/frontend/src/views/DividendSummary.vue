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

      <!-- 按股票分红（最近一年） -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">按股票分红（仅展示最近一年 · 其余折叠）</span>
          <span class="card-sub">{{ displayedStocks.length }} 只有分红数据 · {{ noDataStocks.length }} 只暂无</span>
        </div>

        <div class="stock-div-list">
          <div class="stock-div" v-for="s in displayedStocks" :key="s.code">
            <div class="stock-div__head">
              <span class="stock-div__name">{{ s.name }}</span>
              <span class="stock-div__code">{{ s.code }}</span>
              <span class="stock-div__shares">{{ s.shares }} 股</span>
            </div>

            <div class="stock-div__recent" v-if="s.recent_dividends.length">
              <div class="div-row" v-for="(ev, i) in s.recent_dividends" :key="i">
                <span class="div-date">{{ ev.date }}<em v-if="ev.over_year" class="tag-old">超一年</em></span>
                <span class="div-ratio">每10股派 <b>{{ ev.per_10.toFixed(2) }}</b> 元</span>
                <span class="div-ratio-100">每100股 {{ ev.per_100.toFixed(2) }} 元</span>
              </div>
            </div>
            <div class="stock-div__empty" v-else>暂无近一年分红记录</div>

            <div class="stock-div__older" v-if="s.older_dividends.length">
              <button class="btn-link" @click="toggleOlder(s.code)">
                {{ expanded[s.code] ? '收起' : '展开' }}更早的 {{ s.older_dividends.length }} 笔分红
              </button>
              <div class="older-list" v-show="expanded[s.code]">
                <div class="div-row div-row--muted" v-for="(ev, i) in s.older_dividends" :key="i">
                  <span class="div-date">{{ ev.date }}</span>
                  <span class="div-ratio">每10股派 <b>{{ ev.per_10.toFixed(2) }}</b> 元</span>
                  <span class="div-ratio-100">每100股 {{ ev.per_100.toFixed(2) }} 元</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 无分红数据的股票（沉底，悬停查看） -->
      <div class="card no-data-card" v-if="noDataStocks.length">
        <div class="card-header">
          <span class="card-title">以下股票暂无分红数据（鼠标滑过查看）</span>
        </div>
        <div class="no-data-list">
          <span
            class="no-data-chip"
            v-for="s in noDataStocks"
            :key="s.code"
            @mouseenter="hovered = s"
            @mouseleave="hovered = null"
          >{{ s.name }}</span>
        </div>
        <div class="no-data-tip" v-if="hovered">
          <div class="tip-name">{{ hovered.name }} <span class="tip-code">{{ hovered.code }}</span></div>
          <div class="tip-reason">{{ hovered.shares }} 股 · 暂无分红记录 / 数据缺失，无法展示分红比例</div>
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
        by_stock: [],
        total_annual_dividend: 0,
        total_market_value: 0,
        total_holdings: 0
      },
      expanded: {},
      hovered: null,
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
    },
    // 有数据的排前面
    displayedStocks() {
      return (this.summary.by_stock || []).filter(s => s.has_data)
    },
    // 展示不了（无数据）的沉到最下方
    noDataStocks() {
      return (this.summary.by_stock || []).filter(s => !s.has_data)
    }
  },
  methods: {
    formatNumber(num) {
      return Number(num || 0).toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
    },
    toggleOlder(code) {
      this.$set(this.expanded, code, !this.expanded[code])
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

.card-sub {
  font-size: 12px;
  color: var(--text-muted);
  font-weight: 400;
}

.text-orange { color: #e8590c; }

/* 按股票分红列表 */
.stock-div-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.stock-div {
  border: 1px solid var(--border-color, #e0e0e0);
  border-radius: 10px;
  padding: 10px 14px;
  background: var(--card-bg, #fff);
}

.stock-div__head {
  display: flex;
  align-items: baseline;
  gap: 10px;
  margin-bottom: 6px;
}

.stock-div__name {
  font-weight: 600;
  font-size: 14px;
  color: var(--text-primary, #202124);
}

.stock-div__code {
  font-size: 12px;
  color: var(--text-muted);
}

.stock-div__shares {
  margin-left: auto;
  font-size: 12px;
  color: var(--text-secondary, #5f6368);
}

.stock-div__recent { display: flex; flex-direction: column; gap: 4px; }

.div-row {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 13px;
  padding: 3px 0;
}

.div-row--muted { color: var(--text-muted); }

.div-date {
  min-width: 96px;
  color: var(--text-secondary, #5f6368);
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.tag-old {
  font-style: normal;
  font-size: 10px;
  color: #b45309;
  background: #fef3c7;
  border-radius: 6px;
  padding: 0 5px;
  line-height: 15px;
}

.div-ratio { color: var(--text-primary, #202124); }
.div-ratio b { color: #e8590c; }

.div-ratio-100 {
  color: var(--text-muted);
  font-size: 12px;
}

.stock-div__empty {
  font-size: 12px;
  color: var(--text-muted);
  padding: 4px 0;
}

.stock-div__older { margin-top: 4px; }

.btn-link {
  background: none;
  border: none;
  color: var(--primary-color, #4f46e5);
  font-size: 12px;
  cursor: pointer;
  padding: 2px 0;
}

.older-list {
  margin-top: 4px;
  border-left: 2px solid var(--border-color, #e0e0e0);
  padding-left: 10px;
}

/* 无数据沉底区 */
.no-data-card { margin-top: 16px; }

.no-data-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 4px 0;
}

.no-data-chip {
  font-size: 12px;
  padding: 4px 10px;
  border-radius: 14px;
  background: #f1f5f9;
  color: var(--text-secondary, #5f6368);
  cursor: default;
  border: 1px solid transparent;
  transition: border-color .15s, background .15s;
}

.no-data-chip:hover {
  border-color: var(--primary-color, #4f46e5);
  background: #eef2ff;
}

.no-data-tip {
  margin-top: 10px;
  padding: 10px 14px;
  border-radius: 8px;
  background: #f8fafc;
  border: 1px dashed var(--border-color, #e0e0e0);
}

.tip-name { font-weight: 600; color: var(--text-primary, #202124); }
.tip-code { font-size: 12px; color: var(--text-muted); font-weight: 400; }
.tip-reason { font-size: 12px; color: var(--text-secondary, #5f6368); margin-top: 4px; }
</style>
