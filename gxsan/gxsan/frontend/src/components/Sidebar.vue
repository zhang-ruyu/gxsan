<template>
  <aside class="sidebar">
    <div class="sidebar-header">
      <h1>股息三</h1>
      <div class="version">v1.0.0</div>
    </div>
    <nav class="sidebar-nav">
      <router-link to="/" class="nav-item" exact-active-class="active">
        <span class="icon"> </span>
        <span>分析首页</span>
      </router-link>
      <router-link to="/dashboard" class="nav-item" active-class="active">
        <span class="icon"> </span>
        <span>总览驾驶舱</span>
      </router-link>
      <router-link to="/portfolio" class="nav-item" active-class="active">
        <span class="icon"> </span>
        <span>持仓管理</span>
      </router-link>
      <router-link to="/dividend" class="nav-item" active-class="active">
        <span class="icon"> </span>
        <span>分红汇总</span>
      </router-link>
      <router-link to="/pension" class="nav-item" active-class="active">
        <span class="icon"> </span>
        <span>养老测算</span>
      </router-link>
      <router-link to="/conversion" class="nav-item" active-class="active">
        <span class="icon"> </span>
        <span>资产转换</span>
      </router-link>
      <router-link to="/calendar" class="nav-item" active-class="active">
        <span class="icon"> </span>
        <span>股利日历</span>
      </router-link>
      <router-link to="/settings" class="nav-item" active-class="active">
        <span class="icon">⚙️</span>
        <span>系统设置</span>
      </router-link>
    </nav>
  </aside>
</template>

<script>
export default {
  name: 'Sidebar'
}
</script>
