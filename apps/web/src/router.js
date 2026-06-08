import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from './views/Dashboard.vue'
import Chat from './views/Chat.vue'
import Finance from './views/Finance.vue'
import Health from './views/Health.vue'
import Work from './views/Work.vue'
import Family from './views/Family.vue'
import Notes from './views/Notes.vue'
import Members from './views/Members.vue'
import Settings from './views/Settings.vue'

const routes = [
  { path: '/', name: 'Dashboard', component: Dashboard },
  { path: '/chat', name: 'Chat', component: Chat },
  { path: '/finance', name: 'Finance', component: Finance },
  { path: '/health', name: 'Health', component: Health },
  { path: '/work', name: 'Work', component: Work },
  { path: '/family', name: 'Family', component: Family },
  { path: '/notes', name: 'Notes', component: Notes },
  { path: '/members', name: 'Members', component: Members },
  { path: '/settings', name: 'Settings', component: Settings },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
