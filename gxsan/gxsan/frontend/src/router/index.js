import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/Home.vue'
import Detail from '../views/Detail.vue'
import Portfolio from '../views/Portfolio.vue'
import Calendar from '../views/Calendar.vue'
import Settings from '../views/Settings.vue'
import Tracking from '../views/Tracking.vue'
import DividendSummary from '../views/DividendSummary.vue'
import Dashboard from '../views/Dashboard.vue'

const routes = [
  { path: '/', name: 'Home', component: Home },
  { path: '/dashboard', name: 'Dashboard', component: Dashboard },
  { path: '/detail/:code', name: 'Detail', component: Detail, props: true },
  { path: '/portfolio', name: 'Portfolio', component: Portfolio },
  { path: '/tracking', name: 'Tracking', component: Tracking },
  { path: '/dividend', name: 'DividendSummary', component: DividendSummary },
  { path: '/calendar', name: 'Calendar', component: Calendar },
  { path: '/settings', name: 'Settings', component: Settings }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
