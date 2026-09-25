type SEOInput = {
  title: string
  description: string
  path?: string
  type?: 'website' | 'article'
}

function upsertMeta(attribute: 'name' | 'property', key: string, content: string) {
  let element = document.head.querySelector<HTMLMetaElement>(`meta[${attribute}="${key}"]`)
  if (!element) {
    element = document.createElement('meta')
    element.setAttribute(attribute, key)
    document.head.appendChild(element)
  }
  element.content = content
}

export function setPageSEO(input: SEOInput) {
  if (typeof document === 'undefined') return
  const canonical = new URL(input.path || window.location.pathname, window.location.origin).toString()
  document.title = input.title
  upsertMeta('name', 'description', input.description)
  upsertMeta('property', 'og:title', input.title)
  upsertMeta('property', 'og:description', input.description)
  upsertMeta('property', 'og:type', input.type || 'website')
  upsertMeta('property', 'og:url', canonical)
  let link = document.head.querySelector<HTMLLinkElement>('link[rel="canonical"]')
  if (!link) {
    link = document.createElement('link')
    link.rel = 'canonical'
    document.head.appendChild(link)
  }
  link.href = canonical
}
