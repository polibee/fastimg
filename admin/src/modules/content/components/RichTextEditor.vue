<script setup lang="ts">
import type { JSONContent } from '@tiptap/core'
import { onBeforeUnmount, watch } from 'vue'
import { Bold, Code, Heading2, Image as ImageIcon, Italic, Link as LinkIcon, List, ListOrdered, Quote, Redo2, Undo2 } from '@lucide/vue'
import { EditorContent, useEditor } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Link from '@tiptap/extension-link'
import Image from '@tiptap/extension-image'
import { Button } from '@/components/ui/button'

const props = withDefaults(defineProps<{ modelValue: JSONContent; disabled?: boolean }>(), { disabled: false })
const emit = defineEmits<{ (event: 'update:modelValue', value: JSONContent): void }>()

const editor = useEditor({
  content: props.modelValue,
  editable: !props.disabled,
  extensions: [
    StarterKit,
    Link.configure({ openOnClick: false, autolink: true, protocols: ['http', 'https', 'mailto'] }),
    Image.configure({ inline: false, allowBase64: false }),
  ],
  onUpdate: ({ editor: instance }) => emit('update:modelValue', instance.getJSON()),
})

watch(() => props.disabled, (disabled) => editor.value?.setEditable(!disabled))
watch(() => props.modelValue, (value) => {
  const instance = editor.value
  if (!instance || JSON.stringify(instance.getJSON()) === JSON.stringify(value)) return
  instance.commands.setContent(value, { emitUpdate: false })
})

function setLink() {
  const href = window.prompt('URL')?.trim()
  if (!href) return
  editor.value?.chain().focus().setLink({ href }).run()
}

function addImage() {
  const src = window.prompt('Image URL')?.trim()
  if (!src) return
  editor.value?.chain().focus().setImage({ src, alt: '' }).run()
}

onBeforeUnmount(() => editor.value?.destroy())
</script>

<template>
  <div class="overflow-hidden rounded-lg border bg-background">
    <div class="flex flex-wrap items-center gap-1 border-b bg-muted/30 p-1">
      <Button type="button" variant="ghost" size="icon" class="size-8" :disabled="disabled" :aria-label="'Bold'" @click="editor?.chain().focus().toggleBold().run()"><Bold class="size-4" /></Button>
      <Button type="button" variant="ghost" size="icon" class="size-8" :disabled="disabled" :aria-label="'Italic'" @click="editor?.chain().focus().toggleItalic().run()"><Italic class="size-4" /></Button>
      <Button type="button" variant="ghost" size="icon" class="size-8" :disabled="disabled" :aria-label="'Heading'" @click="editor?.chain().focus().toggleHeading({ level: 2 }).run()"><Heading2 class="size-4" /></Button>
      <Button type="button" variant="ghost" size="icon" class="size-8" :disabled="disabled" :aria-label="'Bullet list'" @click="editor?.chain().focus().toggleBulletList().run()"><List class="size-4" /></Button>
      <Button type="button" variant="ghost" size="icon" class="size-8" :disabled="disabled" :aria-label="'Ordered list'" @click="editor?.chain().focus().toggleOrderedList().run()"><ListOrdered class="size-4" /></Button>
      <Button type="button" variant="ghost" size="icon" class="size-8" :disabled="disabled" :aria-label="'Quote'" @click="editor?.chain().focus().toggleBlockquote().run()"><Quote class="size-4" /></Button>
      <Button type="button" variant="ghost" size="icon" class="size-8" :disabled="disabled" :aria-label="'Code'" @click="editor?.chain().focus().toggleCode().run()"><Code class="size-4" /></Button>
      <Button type="button" variant="ghost" size="icon" class="size-8" :disabled="disabled" :aria-label="'Link'" @click="setLink"><LinkIcon class="size-4" /></Button>
      <Button type="button" variant="ghost" size="icon" class="size-8" :disabled="disabled" :aria-label="'Image'" @click="addImage"><ImageIcon class="size-4" /></Button>
      <span class="mx-1 h-5 w-px bg-border" />
      <Button type="button" variant="ghost" size="icon" class="size-8" :disabled="disabled" :aria-label="'Undo'" @click="editor?.chain().focus().undo().run()"><Undo2 class="size-4" /></Button>
      <Button type="button" variant="ghost" size="icon" class="size-8" :disabled="disabled" :aria-label="'Redo'" @click="editor?.chain().focus().redo().run()"><Redo2 class="size-4" /></Button>
    </div>
    <EditorContent :editor="editor" class="content-editor min-h-64 px-4 py-3" />
  </div>
</template>

<style scoped>
:deep(.ProseMirror) { min-height: 15rem; outline: none; }
:deep(.ProseMirror p) { margin: .55rem 0; }
:deep(.ProseMirror h2) { margin: 1rem 0 .5rem; font-size: 1.25rem; font-weight: 600; }
:deep(.ProseMirror ul) { list-style: disc; padding-left: 1.5rem; }
:deep(.ProseMirror ol) { list-style: decimal; padding-left: 1.5rem; }
:deep(.ProseMirror blockquote) { border-left: 3px solid hsl(var(--border)); padding-left: 1rem; color: hsl(var(--muted-foreground)); }
:deep(.ProseMirror a) { color: hsl(var(--primary)); text-decoration: underline; }
:deep(.ProseMirror img) { max-width: 100%; border-radius: .5rem; }
</style>
