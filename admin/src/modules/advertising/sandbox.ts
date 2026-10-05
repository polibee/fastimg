function containsHtmlTag(source: string) {
  return /<\s*\/?\s*[A-Za-z][\w:-]*(?:\s+[^<>]*)?\s*\/?>/i.test(source)
}

const frameStyles = '<style>html,body{box-sizing:border-box;margin:0;max-width:100%;overflow:hidden;padding:0;background:transparent}*,*::before,*::after{box-sizing:inherit}img,video,iframe,object{display:block;max-width:100%;height:auto}</style>'

export function buildSandboxedAdDocument(source: string) {
  const normalized = source.trim()
  // Complete embeds from any advertising provider stay inside the sandboxed
  // iframe. This intentionally does not identify or whitelist a vendor.
  if (containsHtmlTag(normalized)) return `<!doctype html><html><head>${frameStyles}</head><body>${normalized}</body></html>`

  const escaped = source.replace(/<\/script/gi, '<\\/script')
  return `<!doctype html><html><head>${frameStyles}</head><body><script>${escaped}<\/script></body></html>`
}
