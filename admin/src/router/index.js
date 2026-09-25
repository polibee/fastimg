import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { adminLegacyResourceRedirects, adminRoutes } from '@/apps/admin/routes';
import { memberRoutes } from '@/apps/member/routes';
import { hasAdminAccess } from '@/lib/admin-access';
const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: '/login',
            name: 'login',
            component: () => import('@/modules/auth/pages/LoginPage.vue'),
        },
        {
            path: '/app/media',
            redirect: (to) => ({ path: '/media', query: to.query, hash: to.hash }),
        },
        {
            path: '/app/plans',
            redirect: (to) => ({ path: '/plans', query: to.query, hash: to.hash }),
        },
        {
            path: '/app',
            redirect: (to) => ({ path: '/', query: to.query, hash: to.hash }),
        },
        { path: '/rbac', redirect: '/admin/rbac' },
        { path: '/audit-logs', redirect: '/admin/audit-logs' },
        { path: '/users', redirect: '/admin/users' },
        { path: '/users/new', redirect: '/admin/users/new' },
        { path: '/users/:id/edit', redirect: (to) => ({ path: `/admin${to.path}`, query: to.query, hash: to.hash }) },
        { path: '/roles', redirect: '/admin/roles' },
        { path: '/roles/new', redirect: '/admin/roles/new' },
        { path: '/roles/:id/edit', redirect: (to) => ({ path: `/admin${to.path}`, query: to.query, hash: to.hash }) },
        { path: '/permissions', redirect: '/admin/permissions' },
        { path: '/plans/new', redirect: '/admin/plans/new' },
        { path: '/plans/:id', redirect: (to) => ({ path: `/admin${to.path}`, query: to.query, hash: to.hash }) },
        { path: '/plans/:id/edit', redirect: (to) => ({ path: `/admin${to.path}`, query: to.query, hash: to.hash }) },
        ...adminLegacyResourceRedirects,
        memberRoutes,
        adminRoutes,
        { path: '/forbidden', name: 'forbidden', component: () => import('@/core/pages/ForbiddenPage.vue') },
    ],
});
router.beforeEach(async (to) => {
    const auth = useAuthStore();
    await auth.restore();
    if (to.meta.requiresAuth && !auth.isAuthenticated) {
        return { name: 'login', query: { redirect: to.fullPath } };
    }
    if (to.name === 'login' && auth.isAuthenticated) {
        return { path: '/' };
    }
    if (to.meta.requiresAdminAccess && !hasAdminAccess(auth.user?.permissions || [])) {
        return { name: 'forbidden' };
    }
    if (to.meta.permission && !auth.can(String(to.meta.permission))) {
        return { name: 'forbidden' };
    }
    if (to.meta.anyPermissions && !auth.canAny(to.meta.anyPermissions)) {
        return { name: 'forbidden' };
    }
    const resource = typeof to.params.resource === 'string' ? to.params.resource : '';
    const resourcePermission = { users: 'admin.users.view', roles: 'admin.roles.manage', permissions: 'admin.permissions.manage', plans: 'admin.plans.view' };
    if (resource && resourcePermission[resource] && !auth.can(resourcePermission[resource])) {
        return { name: 'forbidden' };
    }
});
export default router;
