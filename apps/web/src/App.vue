<template>
  <div id="app-container">
    <nav class="navbar">
      <div class="nav-brand">
        <router-link to="/">LifeOps</router-link>
      </div>
      <button class="nav-toggle" @click="menuOpen = !menuOpen">
        <span></span><span></span><span></span>
      </button>
      <ul class="nav-links" :class="{ open: menuOpen }">
        <li><router-link to="/" @click="menuOpen = false">仪表盘</router-link></li>
        <li><router-link to="/chat" @click="menuOpen = false">AI 对话</router-link></li>
        <li><router-link to="/finance" @click="menuOpen = false">财务</router-link></li>
        <li><router-link to="/health" @click="menuOpen = false">健康</router-link></li>
        <li><router-link to="/movement" @click="menuOpen = false">运动</router-link></li>
        <li><router-link to="/work" @click="menuOpen = false">工作</router-link></li>
        <li><router-link to="/family" @click="menuOpen = false">家庭</router-link></li>
        <li><router-link to="/notes" @click="menuOpen = false">笔记</router-link></li>
        <li><router-link to="/members" @click="menuOpen = false">成员</router-link></li>
        <li><router-link to="/settings" @click="menuOpen = false">设置</router-link></li>
      </ul>
    </nav>
    <main class="main-content">
      <router-view />
    </main>
  </div>
</template>

<script>
export default {
  name: 'App',
  data() {
    return {
      menuOpen: false,
    }
  },
}
</script>
