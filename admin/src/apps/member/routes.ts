import type { RouteRecordRaw } from 'vue-router'
import { memberRouteChildren } from './route-parts'
export const memberRoutes: RouteRecordRaw = { path: '/', component: () => import('@/core/layouts/MemberShell.vue'), children: memberRouteChildren }
