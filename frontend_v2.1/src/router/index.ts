import { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: () => import('@/views/HomeView.vue')
  },
  {
    path: '/chambers',
    name: 'chambers',
    component: () => import('@/views/ChambersView.vue')
  },
  {
    path: '/experiments',
    name: 'experiments',
    component: () => import('@/views/ExperimentsView.vue')
  },
  {
    path: '/experiments/:id',
    name: 'experiment-detail',
    component: () => import('@/views/ExperimentDetailView.vue')
  }
]

export default routes 