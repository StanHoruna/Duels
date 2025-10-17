import {createRouter, createWebHistory} from "vue-router";

const routes = [
  {
    path: "/",
    name: "home",
    component: () => import(/* webpackChunkName: "home" */ "../views/Home.vue"),
  },
  {
    path: "/create",
    name: "create",
    component: () => import(/* webpackChunkName: "create" */ "../views/Create.vue"),
  },
  {
    path: "/duel/:id",
    name: "duel",
    component: () => import(/* webpackChunkName: "duel" */ "../views/DuelDetails.vue"),
  },
  {
    path: "/history",
    name: "history",
    component: () => import(/* webpackChunkName: "history" */ "../views/History.vue"),
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(to, from, savedPosition) {
    if (to.hash) return {selector: to.hash, behavior: 'smooth'}
    return {x: 0, y: 0, behavior: 'smooth'}
  }
});

export default router;
