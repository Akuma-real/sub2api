import DOMPurify from 'dompurify'
import { marked } from 'marked'

export function sanitizeSvg(svg: string): string {
  if (!svg) return ''
  return DOMPurify.sanitize(svg, { USE_PROFILES: { svg: true, svgFilters: true } })
}

export function sanitizeHtml(html: string): string {
  if (!html) return ''
  return DOMPurify.sanitize(html, {
    ADD_ATTR: ['target', 'rel'],
  })
}

export function renderMarkdownToSafeHtml(markdown: string): string {
  if (!markdown.trim()) return ''
  const html = marked.parse(markdown, { async: false }) as string
  return sanitizeHtml(html)
}
