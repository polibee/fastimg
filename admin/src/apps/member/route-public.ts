import type {RouteRecordRaw} from 'vue-router'
const c=(name:string)=>()=>import(`@/modules/member/pages/${name}.vue`)
export const publicMemberRoutes:RouteRecordRaw[]=[{path:'',name:'member-home',component:c('MemberHomePage')},{path:'discover',name:'member-discover',component:c('MemberDiscoverPage')},{path:'plans',name:'member-plans',component:c('MemberPlansPage')}]
