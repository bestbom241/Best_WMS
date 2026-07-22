import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LocationsView from '../views/LocationsView.vue'
import ProductsView from '../views/ProductsView.vue'
import InventoryView from '../views/InventoryView.vue'
import ReportView from '../views/ReportView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: HomeView },
    { path: '/inventory', component: InventoryView },
    { path: '/locations', component: LocationsView },
    { path: '/products', component: ProductsView },
    { path: '/report', component: ReportView }
  ]
})

export default router