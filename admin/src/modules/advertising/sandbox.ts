function containsHtmlTag(source: string) {
  return /<\s*\/?\s*[A-Za-z][\w:-]*(?:\s+[^<>]*)?\s*\/?>/i.test(source)
}

export function buildSandboxedAdDocument(source: string) {
  const normalized = source.trim()
  // Complete embeds from any advertising provider stay inside the sandboxed
  // iframe. This intentionally does not identify or whitelist a vendor.
  if (containsHtmlTag(normalized)) return normalized

  const escaped = source.replace(/<\/script/gi, '<\\/script')
  return `<!doctype html><html><body><script>${escaped}<\/script></body></html>`
}
