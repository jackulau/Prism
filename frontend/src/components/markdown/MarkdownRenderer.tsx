import { useState, memo } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Highlight, themes, type RenderProps } from 'prism-react-renderer'
import { Copy, Check } from 'lucide-react'

interface MarkdownRendererProps {
  content: string
  isStreaming?: boolean
}

// Code block component with syntax highlighting
function CodeBlock({ code, language, isStreaming }: { code: string; language: string; isStreaming?: boolean }) {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    await navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="rounded-lg overflow-hidden border border-editor-border bg-editor-surface my-3">
      <div className="flex items-center justify-between px-4 py-2 bg-editor-bg/50 border-b border-editor-border">
        <span className="text-xs text-editor-muted font-mono">{language || 'text'}</span>
        <button
          onClick={handleCopy}
          className="flex items-center gap-1 text-xs text-editor-muted hover:text-editor-text transition-colors"
        >
          {copied ? (
            <>
              <Check size={12} className="text-editor-success" />
              <span className="text-editor-success">Copied!</span>
            </>
          ) : (
            <>
              <Copy size={12} />
              <span>Copy</span>
            </>
          )}
        </button>
      </div>
      <Highlight theme={themes.nightOwl} code={code.trim()} language={language || 'text'}>
        {({ className, style, tokens, getLineProps, getTokenProps }: RenderProps) => (
          <pre className={`${className} p-4 text-sm font-mono overflow-x-auto`} style={{ ...style, background: 'transparent', margin: 0 }}>
            {tokens.map((line, i) => {
              const lineProps = getLineProps({ line, key: i })
              const isLastLine = i === tokens.length - 1
              return (
                <div key={i} {...lineProps} className="table-row">
                  <span className="table-cell pr-4 text-editor-muted/50 select-none text-right w-8">{i + 1}</span>
                  <span className="table-cell">
                    {line.map((token, key) => (
                      <span key={key} {...getTokenProps({ token, key })} />
                    ))}
                    {isStreaming && isLastLine && (
                      <span className="inline-block w-2 h-4 bg-editor-accent animate-pulse ml-0.5 align-middle" />
                    )}
                  </span>
                </div>
              )
            })}
          </pre>
        )}
      </Highlight>
    </div>
  )
}

// Inline code component
function InlineCode({ children }: { children: React.ReactNode }) {
  return (
    <code className="px-1.5 py-0.5 rounded bg-editor-surface text-editor-accent text-sm font-mono border border-editor-border">
      {children}
    </code>
  )
}

export const MarkdownRenderer = memo(function MarkdownRenderer({ content, isStreaming }: MarkdownRendererProps) {
  // Check if content ends with an incomplete code block (streaming)
  const hasIncompleteCodeBlock = isStreaming && /```[\w]*\n[^`]*$/.test(content)

  return (
    <div className="markdown-content">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          // Headers
          h1: ({ children }) => (
            <h1 className="text-2xl font-bold mt-6 mb-4 text-editor-text border-b border-editor-border pb-2">
              {children}
            </h1>
          ),
          h2: ({ children }) => (
            <h2 className="text-xl font-bold mt-5 mb-3 text-editor-text border-b border-editor-border pb-2">
              {children}
            </h2>
          ),
          h3: ({ children }) => (
            <h3 className="text-lg font-semibold mt-4 mb-2 text-editor-text">{children}</h3>
          ),
          h4: ({ children }) => (
            <h4 className="text-base font-semibold mt-3 mb-2 text-editor-text">{children}</h4>
          ),
          h5: ({ children }) => (
            <h5 className="text-sm font-semibold mt-3 mb-1 text-editor-text">{children}</h5>
          ),
          h6: ({ children }) => (
            <h6 className="text-sm font-semibold mt-3 mb-1 text-editor-muted">{children}</h6>
          ),

          // Paragraphs
          p: ({ children }) => (
            <p className="my-2 leading-relaxed">{children}</p>
          ),

          // Lists
          ul: ({ children }) => (
            <ul className="list-disc list-inside my-3 space-y-1.5 pl-2">{children}</ul>
          ),
          ol: ({ children }) => (
            <ol className="list-decimal list-inside my-3 space-y-1.5 pl-2">{children}</ol>
          ),
          li: ({ children }) => (
            <li className="leading-relaxed">{children}</li>
          ),

          // Blockquotes
          blockquote: ({ children }) => (
            <blockquote className="border-l-4 border-editor-accent pl-4 my-4 italic text-editor-muted bg-editor-surface/30 py-2 rounded-r">
              {children}
            </blockquote>
          ),

          // Tables
          table: ({ children }) => (
            <div className="overflow-x-auto my-4">
              <table className="min-w-full border border-editor-border rounded-lg overflow-hidden">
                {children}
              </table>
            </div>
          ),
          thead: ({ children }) => (
            <thead className="bg-editor-surface">{children}</thead>
          ),
          tbody: ({ children }) => (
            <tbody className="divide-y divide-editor-border">{children}</tbody>
          ),
          tr: ({ children }) => (
            <tr className="hover:bg-editor-surface/50 transition-colors">{children}</tr>
          ),
          th: ({ children }) => (
            <th className="px-4 py-3 text-left text-sm font-semibold text-editor-text border-b border-editor-border">
              {children}
            </th>
          ),
          td: ({ children }) => (
            <td className="px-4 py-3 text-sm">{children}</td>
          ),

          // Code blocks
          code: ({ className, children }) => {
            // Check if inline by looking at the node or parent - inline code doesn't have className
            const isInline = !className && !String(children).includes('\n')

            if (isInline) {
              return <InlineCode>{children}</InlineCode>
            }

            const match = /language-(\w+)/.exec(className || '')
            const language = match ? match[1] : ''
            const codeString = String(children).replace(/\n$/, '')

            // Check if this is the last code block and we're streaming
            const isLastCodeBlock = isStreaming && hasIncompleteCodeBlock

            return (
              <CodeBlock
                code={codeString}
                language={language}
                isStreaming={isLastCodeBlock}
              />
            )
          },
          pre: ({ children }) => <>{children}</>,

          // Links
          a: ({ href, children }) => (
            <a
              href={href}
              target="_blank"
              rel="noopener noreferrer"
              className="text-editor-accent hover:underline"
            >
              {children}
            </a>
          ),

          // Horizontal rules
          hr: () => <hr className="my-6 border-editor-border" />,

          // Strong/Bold
          strong: ({ children }) => (
            <strong className="font-semibold text-editor-text">{children}</strong>
          ),

          // Emphasis/Italic
          em: ({ children }) => (
            <em className="italic">{children}</em>
          ),

          // Strikethrough
          del: ({ children }) => (
            <del className="line-through text-editor-muted">{children}</del>
          ),

          // Images
          img: ({ src, alt }) => (
            <img
              src={src}
              alt={alt}
              className="max-w-full h-auto rounded-lg my-4 border border-editor-border"
            />
          ),
        }}
      >
        {content}
      </ReactMarkdown>

      {/* Streaming cursor for non-code content */}
      {isStreaming && !hasIncompleteCodeBlock && (
        <span className="inline-block w-2 h-4 bg-editor-accent animate-pulse ml-0.5 align-middle" />
      )}
    </div>
  )
})

export default MarkdownRenderer
