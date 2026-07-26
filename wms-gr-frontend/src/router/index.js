import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LocationsView from '../views/LocationsView.vue'
import ProductsView from '../views/ProductsView.vue'
import InventoryView from '../views/InventoryView.vue'
import ReportView from '../views/ReportView.vue'
import SuppliersView from '../views/SuppliersView.vue'
import CustomersView from '../views/CustomersView.vue'
import GRPlansView from '../views/GRPlansView.vue'
import OutboundPlansView from '../views/OutboundPlansView.vue'
import PickingView from '../views/PickingView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: HomeView },
    { path: '/gr-plans', component: GRPlansView },
    { path: '/outbound-plans', component: OutboundPlansView },
    { path: '/picking', component: PickingView },
    { path: '/inventory', component: InventoryView },
    { path: '/locations', component: LocationsView },
    { path: '/products', component: ProductsView },
    { path: '/suppliers', component: SuppliersView },
    { path: '/customers', component: CustomersView },
    { path: '/report', component: ReportView }
  ]
})

export default router
