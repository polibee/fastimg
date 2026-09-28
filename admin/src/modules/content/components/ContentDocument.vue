<script setup lang="ts">
import type { JSONContent } from '@tiptap/core'
import { h, type VNode } from 'vue'

defineOptions({ name: 'ContentDocument' })
const props = defineProps<{ document: JSONContent }>()

function renderNode(node: JSONContent): VNode | string | null {
  if (node.type === 'text') {
    let value: VNode | string = node.text || ''
    for (const mark of [...(node.marks || [])].reverse()) {
      if (mark.type === 'bold') value = h('strong', {}, [value])
      else if (mark.type === 'italic') value = h('em', {}, [value])
      else if (mark.type === 'strike') value = h('s', {}, [value])
      else if (mark.type === 'code') value = h('code', { class: 'rounded bg-muted px-1 py-0.5 text-sm' }, [value])
      else if (mark.type === 'link' && typeof mark.attrs?.href === 'string') value = h('a', { href: mark.attrs.href, target: '_blank', rel: 'noopener noreferrer', class: 'text-primary underline' }, [value])
    }
    return value
  }
  const children = (node.content || []).map(renderNode).filter(Boolean) as (VNode | string)[]
  if (node.type === 'doc') return h('div', {}, children)
  if (node.type === 'paragraph') return h('p', { class: 'my-3' }, children)
  if (node.type === 'heading') return h(`h${Math.min(6, Math.max(1, Number(node.attrs?.level || 2)))}`, { class: 'my-5 font-semibold' }, children)
  if (node.type === 'bulletList') return h('ul', { class: 'my-3 list-disc space-y-1 pl-6' }, children)
  if (node.type === 'orderedList') return h('ol', { class: 'my-3 list-decimal space-y-1 pl-6' }, children)
  if (node.type === 'listItem') return h('li', {}, children)
  if (node.type === 'blockquote') return h('blockquote', { class: 'my-4 border-l-2 pl-4 text-muted-foreground' }, children)
  if (node.type === 'codeBlock') return h('pre', { class: 'my-4 overflow-x-auto rounded-lg bg-muted p-4 text-sm' }, [h('code', {}, children)])
  if (node.type === 'hardBreak') return h('br')
  if (node.type === 'horizontalRule') return h('hr', { class: 'my-6' })
  if (node.type === 'image' && typeof node.attrs?.src === 'string') return h('img', { src: node.attrs.src, alt: node.attrs.alt || '', loading: 'lazy', class: 'my-4 max-w-full rounded-lg' })
  return children.length ? h('div', {}, children) : null
}

const renderDocument = () => renderNode(props.document)
</script>

<template><component :is="renderDocument" /></template>
