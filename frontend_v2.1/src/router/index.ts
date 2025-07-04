import { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/LoginView.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('@/views/RegisterView.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    name: 'home',
    component: () => import('@/views/HomeView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/profile',
    name: 'profile',
    component: () => import('@/views/ProfileView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/chambers',
    name: 'chambers',
    component: () => import('@/views/ChambersView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/experiments',
    name: 'experiments',
    component: () => import('@/views/ExperimentsView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/experiments/:id',
    name: 'experiment-detail',
    component: () => import('@/views/ExperimentDetailView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/api-tokens',
    name: 'api-tokens',
    component: () => import('@/views/ApiTokensView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/admin/users',
    name: 'admin-users',
    component: () => import('@/views/AdminUsersView.vue'),
    meta: { requiresAuth: true, requiresAdmin: true }
  },
  {
    path: '/chambers/:id/config',
    name: 'chamber-config',
    component: () => import('@/views/ChamberConfigView.vue'),
    meta: { requiresAuth: true }
  },
]

export default routes