<template>
  <div>
    <div class="page-header">
      <h2>分析首页</h2>
      <div class="header-actions">
        <span class="last-update" v-if="lastUpdate">最后更新: {{ lastUpdate }}</span>
        <button class="btn btn-secondary btn-sm" @click="refreshData">
          刷新数据
        </button>
      </div>
    </div>

    <div v-if="loading" class="loading">
      <div class="loading-spinner"></div>
      <p>正在获取数据并分析...</p>
    </div>

    <div v-else>
      <!-- 汇总统计 -->
      <div class="summary-stats">
        <div class="stat-card">
          <div class="stat-value">{{ watchlist.length }}</div>
          <div class="stat-label">监控股票</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ buyCount }}</div>
          <div class="stat-label">买入信号</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">¥{{ formatNumber(fundInfo.total_assets) }}</div>
          <div class="stat-label">总资产</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">¥{{ formatNumber(fundInfo.available_fund) }}</div>
          <div class="stat-label">可用资金</div>
        </div>
      </div>

      <!-- 股票列表 -->
      <div class="stock-list">
        <StockCard
          v-for="stock in watchlist"
          :key="stock.code"
          :stock="stock"
          @click="goToDetail(stock.code)"
        />
      </div>

      <div v-if="watchlist.length === 0" class="empty-state">
        <div class="icon"> </div>
        <p>暂无监控股票</p>
        <p>请在设置中添加监控股票</p>
      </div>
    </div>
  </div>
</template>

<script>
import { GetWatchlist, GetFundInfo } from '../../wailsjs/go/main/App'
import StockCard from '../components/StockCard.vue'

export default {
  name: 'Home',
  components: {
    StockCard
  },
  data() {
    return {
      loading: true,
      watchlist: [],
      fundInfo: {
        available_fund: 0,
        total_market: 0,
        total_assets: 0
      },
      lastUpdate: '',
      refreshTimer: null
    }
  },
  computed: {
    buyCount() {
      return this.watchlist.filter(s => s.signal === 'BUY').length
    }
  },
  methods: {
    formatNumber(num) {
      return num.toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
    },
    goToDetail(code) {
      this.$router.push(`/detail/${code}`)
    },
    async refreshData() {
      try {
        console.log('开始刷新数据...')
        const watchlistStr = await GetWatchlist()
        console.log('获取到监控列表:', watchlistStr)
        this.watchlist = JSON.parse(watchlistStr)
        
        const fundStr = await GetFundInfo()
        console.log('获取到资金信息:', fundStr)
        this.fundInfo = JSON.parse(fundStr)
        
        this.lastUpdate = new Date().toLocaleTimeString('zh-CN')
        console.log('刷新完成')
      } catch (error) {
        console.error('获取数据失败:', error)
        alert('刷新失败: ' + error.message)
      }
      this.loading = false
    },
    startAutoRefresh() {
      if (this.refreshTimer) {
        clearInterval(this.refreshTimer)
      }
      this.refreshTimer = setInterval(() => {
        console.log('自动刷新触发')
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
    console.log('Home组件挂载')
    this.refreshData()
    this.startAutoRefresh()
  },
  beforeUnmount() {
    this.stopAutoRefresh()
  }
}
</script>

<style scoped>
.stock-list {
  display: grid;
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
</style>
