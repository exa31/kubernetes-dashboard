import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: () => import('@/layouts/DashboardLayout.vue'),
      children: [
        {
          path: '',
          redirect: '/overview',
        },
        {
          path: 'overview',
          name: 'ClusterOverview',
          component: () => import('@/pages/ClusterOverviewPage.vue'),
        },
        {
          path: 'secrets',
          name: 'Secrets',
          component: () => import('@/pages/SecretsPage.vue'),
        },
        {
          path: 'configmaps',
          name: 'ConfigMaps',
          component: () => import('@/pages/ConfigMapsPage.vue'),
        },
        {
          path: 'workloads',
          name: 'Workloads',
          component: () => import('@/pages/WorkloadsPage.vue'),
        },
        {
          path: 'services',
          name: 'Services',
          component: () => import('@/pages/ServicesPage.vue'),
        },
        {
          path: 'ingresses',
          name: 'Ingresses',
          component: () => import('@/pages/IngressesPage.vue'),
        },
        {
          path: 'cronjobs',
          name: 'CronJobs',
          component: () => import('@/pages/CronJobsPage.vue'),
        },
        {
          path: 'storage',
          name: 'Storage',
          component: () => import('@/pages/StoragePage.vue'),
        },
        {
          path: 'users',
          name: 'Users',
          component: () => import('@/pages/Users.vue'),
        },
      ],
    },
    {
      path: '/auth',
      component: () => import('@/layouts/AuthLayout.vue'),
      children: [
        {
          path: 'login',
          name: 'Login',
          component: () => import('@/pages/Login.vue')
        }
      ]
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'NotFound',
      component: () => import('@/pages/NotFound.vue')
    }
  ]
})

export default router
