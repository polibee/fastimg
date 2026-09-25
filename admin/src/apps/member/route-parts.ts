import type { RouteRecordRaw } from 'vue-router'
import { publicMemberRoutes } from './route-public'
import { privateMemberRoutes } from './route-private'
export const memberRouteChildren: RouteRecordRaw[] = [...publicMemberRoutes, ...privateMemberRoutes]
