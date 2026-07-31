<template>
  <div>
    <div class="page-header">
      <h2>股利日历</h2>
      <div>
        <select class="form-input" v-model="days" @change="loadCalendar" style="width: auto; display: inline-block">
          <option :value="30">未来30天</option>
          <option :value="60">未来60天</option>
          <option :value="90">未来90天</option>
        </select>
      </div>
    </div>

    <div v-if="loading" class="loading">
      <div class="loading-spinner"></div>
      <p>加载中...</p>
    </div>

    <div v-else>
      <!-- 汇总信息 -->
      <div class="summary-stats">
        <div class="stat-card">
          <div class="stat-value">{{ portfolioCount }}只</div>
          <div class="stat-label">持仓数量</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">¥{{ annualDividend.toFixed(0) }}</div>
          <div class="stat-label">年化股息总额</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">¥{{ monthlyDividend.toFixed(0) }}</div>
          <div class="stat-label">月均股息</div>
        </div>
      </div>

      <!-- 分红记录 -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">近期分红记录</span>
        </div>
        <table class="table">
          <thead>
            <tr>
              <th>日期</th>
              <th>代码</th>
              <th>名称</th>
              <th>每股股息</th>
              <th>持仓股息</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(record, index) in dividendRecords" :key="index">
              <td>{{ record.date }}</td>
              <td>{{ record.code }}</td>
              <td>{{ record.name }}</td>
              <td>¥{{ record.amount.toFixed(2) }}</td>
              <td>¥{{ record.total.toFixed(0) }}</td>
            </tr>
          </tbody>
        </table>

        <div v-if="dividendRecords.length === 0" class="empty-state">
          <div class="icon"> </div>
          <p>暂无分红记录</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { GetCalendar } from '../../wailsjs/go/main/App'

export default {
  name: 'Calendar',
  data() {
    return {
      loading: true,
      days: 30,
      dividendRecords: [],
      portfolioCount: 0,
      annualDividend: 0,
      monthlyDividend: 0
    }
  },
  methods: {
    async loadCalendar() {
      this.loading = true
      try {
        const data = await GetCalendar(this.days)
        this.parseCalendar(data)
      } catch (error) {
        console.error('获取日历失败:', error)
      }
      this.loading = false
    },
    parseCalendar(text) {
      const lines = text.split('\n')
      this.dividendRecords = []
      
      for (const line of lines) {
        if (line.includes('持仓数量')) {
          const match = line.match(/持仓数量:\s*(\d+)/)
          if (match) this.portfolioCount = parseInt(match[1])
        }
        if (line.includes('年化股息总额')) {
          const match = line.match(/年化股息总额:\s*¥([\d.]+)/)
          if (match) this.annualDividend = parseFloat(match[1])
        }
        if (line.includes('月均股息')) {
          const match = line.match(/月均股息:\s*¥([\d.]+)/)
          if (match) this.monthlyDividend = parseFloat(match[1])
        }
        
        // 解析分红记录行
        const recordMatch = line.match(/(\d{4}-\d{2}-\d{2})\s+(\d+)\s+(\S+)\s+¥([\d.]+)\s+¥([\d.]+)/)
        if (recordMatch) {
          this.dividendRecords.push({
            date: recordMatch[1],
            code: recordMatch[2],
            name: recordMatch[3],
            amount: parseFloat(recordMatch[4]),
            total: parseFloat(recordMatch[5])
          })
        }
      }
    }
  },
  mounted() {
    this.loadCalendar()
  }
}
</script>
