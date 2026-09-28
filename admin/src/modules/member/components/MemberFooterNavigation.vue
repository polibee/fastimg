<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { footerPublicApi, type NavigationGroupResult, type NavigationItem } from '@/modules/footer-navigation/api'
import { useI18n } from 'vue-i18n'
const { locale } = useI18n(); const groups = ref<NavigationGroupResult[]>([])
function itemHref(item: NavigationItem) { if (item.target_type === 'page') return `/page/${encodeURIComponent(item.target_value)}`; if (item.target_type === 'friends') return '/friends'; return item.target_value }
function children(group: NavigationGroupResult, parentID: number) { return group.items.filter((item) => item.parent_id === parentID) }
async function load() { try { groups.value = await footerPublicApi.list(locale.value) } catch { groups.value = [] } }
onMounted(load)
watch(locale, load)
</script>

<template><footer v-if="groups.length" class="border-t bg-card"><div class="mx-auto grid max-w-6xl gap-8 px-4 py-8 sm:grid-cols-2 lg:grid-cols-4 sm:px-6"><section v-for="entry in groups" :key="entry.group.id"><h2 class="text-sm font-semibold">{{ entry.group.title }}</h2><ul class="mt-3 space-y-2 text-sm text-muted-foreground"><li v-for="item in entry.items.filter((value) => !value.parent_id)" :key="item.id"><a :href="itemHref(item)" :target="item.open_in_new_tab ? '_blank' : undefined" :rel="item.open_in_new_tab ? 'noopener noreferrer' : undefined" class="hover:text-foreground hover:underline">{{ item.label }}</a><ul v-if="children(entry, item.id).length" class="mt-2 space-y-2 border-l pl-3"><li v-for="child in children(entry, item.id)" :key="child.id"><a :href="itemHref(child)" :target="child.open_in_new_tab ? '_blank' : undefined" :rel="child.open_in_new_tab ? 'noopener noreferrer' : undefined" class="hover:text-foreground hover:underline">{{ child.label }}</a></li></ul></li></ul></section></div></footer></template>
