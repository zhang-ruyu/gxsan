<template>
  <div>
    <div class="page-header">
      <h2>跟踪推荐</h2>
      <div class="header-actions">
        <span class="last-update" v-if="lastUpdate">最后更新: {{ lastUpdate }}</span>
        <span class="refresh-state" v-if="refreshState">{{ refreshState }}</span>
        <button class="btn btn-secondary btn-sm" @click="refreshData">刷新数据</button>
      </div>
    </div>

    <div v-if="loading" class="loading">
      <div class="loading-spinner"></div>
      <p>正在加载推荐股票池...</p>
    </div>

    <div v-else-if="error" class="error-state">
      <p>❌ {{ error }}</p>
      <button class="btn btn-primary" @click="refreshData">重试</button>
    </div>

    <div v-else>
      <!-- 数据截止提示 -->
      <div class="data-notice">
        <span class="notice-icon"> </span>
        <span>静态数据（股息率/逻辑/风险/建议）截止 <strong>2026年7月25日</strong>，源自 skill 推荐池。实时行情来自东方财富，交易时段每 30 秒自动刷新。</span>
      </div>

      <!-- 汇总统计 -->
      <div class="summary-stats">
        <div class="stat-card">
          <div class="stat-value">{{ totalStocks }}</div>
          <div class="stat-label">推荐标的</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ categories.length }}</div>
          <div class="stat-label">分类</div>
        </div>
        <div class="stat-card">
          <div class="stat-value text-green">{{ undervaluedCount }}</div>
          <div class="stat-label">实时股息率 ≥ skill 值</div>
        </div>
        <div class="stat-card">
          <div class="stat-value text-orange">{{ liveCount }}</div>
          <div class="stat-label">已获取实时行情</div>
        </div>
      </div>

      <!-- 分类筛选 -->
      <div class="filter-bar">
        <button class="btn btn-sm" :class="activeCategory === 'all' ? 'btn-primary' : 'btn-secondary'"
                @click="activeCategory = 'all'">全部</button>
        <button v-for="cat in categories" :key="cat.name"
                class="btn btn-sm"
                :class="activeCategory === cat.name ? 'btn-primary' : 'btn-secondary'"
                @click="activeCategory = cat.name">{{ cat.name }} ({{ cat.stocks.length }})</button>
      </div>

      <!-- 分类区域 -->
      <div v-for="cat in displayedCategories" :key="cat.name" class="card category-card">
        <div class="card-header" @click="toggleCategory(cat.name)">
          <span class="card-title">{{ cat.name }}</span>
          <span class="category-count">{{ cat.stocks.length }} 只</span>
          <span class="toggle-icon" :class="{ expanded: expandedCategories.has(cat.name) }">▼</span>
        </div>

        <div v-if="expandedCategories.has(cat.name)" class="category-body">
          <div class="stock-grid">
            <div v-for="stock in cat.stocks" :key="stock.code"
                 class="stock-card"
                 :class="{ 'undervalued': isUndervalued(stock), 'clickable': true }"
                 @click="goToDetail(stock.code)">
              <!-- 卡片头部 -->
              <div class="stock-header">
                <div class="stock-name">{{ stock.name }}</div>
                <div class="stock-code">{{ stock.code }}</div>
              </div>

              <!-- 类型标签 -->
              <div class="stock-badges">
                <span class="badge" :class="stock.asset_type === '弱周期' ? 'badge-blue' : 'badge-orange'">
                  {{ stock.asset_type }}
                </span>
                <span class="badge badge-gray">skill: {{ stock.skill_yield }}</span>
              </div>

              <!-- 实时行情 -->
              <div class="stock-price-row" v-if="stock.price > 0">
                <span class="price-label">现价</span>
                <span class="price-value">¥{{ formatPrice(stock.price) }}</span>
                <span class="yield-label">实时股息率</span>
                <span class="yield-value" :class="yieldClass(stock)">
                  {{ stock.current_yield.toFixed(2) }}%
                </span>
              </div>
              <div v-else class="no-data">暂无实时行情</div>

              <!-- 分红信息 -->
              <div class="stock-dps" v-if="stock.dividend_per_share > 0">
                每股分红 ¥{{ stock.dividend_per_share.toFixed(4) }}
              </div>

              <!-- 核心逻辑 -->
              <div class="stock-info">
                <div class="info-row">
                  <span class="info-label">逻辑</span>
                  <span class="info-value">{{ stock.core_logic }}</span>
                </div>
                <div class="info-row">
                  <span class="info-label">风险</span>
                  <span class="info-value text-red">{{ stock.key_risk }}</span>
                </div>
                <div class="info-row">
                  <span class="info-label">建议</span>
                  <span class="info-value text-orange">{{ stock.advice }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 底部说明 -->
      <div class="footer-note">
        <p>数据来源：微信公众号「分红养老之路」（森哥）实战周记 | 数据截止：2026年7月25日</p>
        <p class="text-muted small">点击股票卡片可进入详情页查看更多分析。实时行情来自东方财富，交易时段每 30 秒自动刷新。</p>
      </div>
    </div>
  </div>
</template>

<script>
import { GetTrackingStocks } from '../../wailsjs/go/main/App'
import { parseJSON } from '../utils/api'
import { isPageVisible, isMarketOpen, refreshReason } from '../utils/market'

export default {
  name: 'Tracking',
  data() {
    return {
      loading: true,
      error: null,
      categories: [],
      activeCategory: 'all',
      expandedCategories: new Set(),
      lastUpdate: '',
      refreshState: '',
      refreshTimer: null
    }
  },
  computed: {
    totalStocks() {
      return this.categories.reduce((sum, cat) => sum + cat.stocks.length, 0)
    },
    liveCount() {
      let n = 0
      for (const cat of this.categories) {
        for (const s of cat.stocks) {
          if (s.price > 0) n++
        }
      }
      return n
    },
    undervaluedCount() {
      let n = 0
      for (const cat of this.categories) {
        for (const s of cat.stocks) {
          if (this.isUndervalued(s)) n++
        }
      }
      return n
    },
    displayedCategories() {
      if (this.activeCategory === 'all') return this.categories
      return this.categories.filter(c => c.name === this.activeCategory)
    }
  },
  methods: {
    formatPrice(num) {
      return Number(num || 0).toFixed(2)
    },
    isUndervalued(stock) {
      if (stock.current_yield <= 0) return false
      const skillNum = parseFloat(stock.skill_yield.replace(/[~%+]/g, ''))
      if (isNaN(skillNum)) return false
      return stock.current_yield >= skillNum
    },
    yieldClass(stock) {
      return this.isUndervalued(stock) ? 'text-green' : ''
    },
    goToDetail(code) {
      this.$router.push(`/detail/${code}`)
    },
    toggleCategory(name) {
      if (this.expandedCategories.has(name)) {
        this.expandedCategories.delete(name)
      } else {
        this.expandedCategories.add(name)
      }
      // 强制响应式更新
      this.expandedCategories = new Set(this.expandedCategories)
    },
    async refreshData() {
      this.error = null
      try {
        this.refreshState = refreshReason()
        const raw = await GetTrackingStocks()
        this.categories = parseJSON(raw)
        // 默认展开所有分类
        this.expandedCategories = new Set(this.categories.map(c => c.name))
        this.lastUpdate = new Date().toLocaleTimeString('zh-CN')
      } catch (error) {
        console.error('获取跟踪推荐失败:', error)
        this.error = error.message || '获取跟踪推荐失败'
      }
      this.loading = false
    },
    startAutoRefresh() {
      if (this.refreshTimer) clearInterval(this.refreshTimer)
      this.refreshTimer = setInterval(() => {
        this.refreshState = refreshReason()
        if (!isPageVisible() || !isMarketOpen()) {
          console.log('跟踪推荐自动刷新跳过:', this.refreshState)
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
.text-red { color: var(--danger-color); }
.text-muted { color: var(--text-muted); }
.small { font-size: 12px; }

.filter-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 16px;
}

.category-card {
  margin-bottom: 16px;
}

.card-header {
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 12px;
}

.category-count {
  font-size: 13px;
  color: var(--text-muted);
}

.toggle-icon {
  margin-left: auto;
  font-size: 12px;
  color: var(--text-muted);
  transition: transform 0.2s ease;
  display: inline-block;
}

.toggle-icon.expanded {
  transform: rotate(180deg);
}

.category-body {
  padding-top: 8px;
}

.stock-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 12px;
}

.stock-card {
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 12px;
  cursor: pointer;
  transition: box-shadow 0.2s ease, border-color 0.2s ease;
}

.stock-card:hover {
  border-color: #6366f1;
  box-shadow: 0 2px 8px rgba(99, 102, 241, 0.15);
}

.stock-card.undervalued {
  border-left: 3px solid var(--success-color);
}

.stock-header {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 8px;
}

.stock-name {
  font-weight: 700;
  font-size: 15px;
}

.stock-code {
  font-size: 12px;
  color: var(--text-muted);
}

.stock-badges {
  display: flex;
  gap: 6px;
  margin-bottom: 8px;
}

.badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
  white-space: nowrap;
}

.badge-blue {
  background: #eef2ff;
  color: #4f46e5;
}

.badge-orange {
  background: #fff4e6;
  color: #e8590c;
}

.badge-gray {
  background: #f1f5f9;
  color: #64748b;
}

.stock-price-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 4px;
  font-size: 13px;
}

.price-label, .yield-label {
  color: var(--text-muted);
  font-size: 12px;
}

.price-value {
  font-weight: 600;
}

.yield-value {
  font-weight: 700;
  margin-left: auto;
}

.stock-dps {
  font-size: 12px;
  color: var(--text-muted);
  margin-bottom: 8px;
}

.no-data {
  font-size: 12px;
  color: var(--text-muted);
  margin-bottom: 8px;
  font-style: italic;
}

.stock-info {
  border-top: 1px solid #f1f5f9;
  padding-top: 8px;
}

.info-row {
  display: flex;
  gap: 8px;
  font-size: 12px;
  line-height: 1.6;
  margin-bottom: 2px;
}

.info-label {
  color: var(--text-muted);
  min-width: 28px;
  flex-shrink: 0;
}

.info-value {
  flex: 1;
}

.data-notice {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  margin-bottom: 16px;
  background: #fffbeb;
  border: 1px solid #fed7aa;
  border-radius: 8px;
  font-size: 13px;
  color: #92400e;
}

.notice-icon {
  font-size: 16px;
}

.footer-note {
  margin-top: 24px;
  padding: 12px 0;
  text-align: center;
  font-size: 12px;
  color: var(--text-muted);
  border-top: 1px solid #e2e8f0;
}

.footer-note p {
  margin: 4px 0;
}
</style>
