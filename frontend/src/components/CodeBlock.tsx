import React, { useMemo } from 'react';
import { Highlight, themes, type RenderProps } from 'prism-react-renderer';
import { Copy, Check, FileCode } from 'lucide-react';

interface CodeBlockProps {
  code: string;
  language?: string;
  filename?: string;
  showLineNumbers?: boolean;
  isStreaming?: boolean;
}

export const CodeBlock: React.FC<CodeBlockProps> = ({
  code,
  language = 'typescript',
  filename,
  showLineNumbers = true,
  isStreaming = false,
}) => {
  const [copied, setCopied] = React.useState(false);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(code);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const normalizedLanguage = useMemo(() => {
    const langMap: Record<string, string> = {
      js: 'javascript',
      ts: 'typescript',
      tsx: 'tsx',
      jsx: 'jsx',
      py: 'python',
      rb: 'ruby',
      sh: 'bash',
      shell: 'bash',
      yml: 'yaml',
      md: 'markdown',
    };
    return langMap[language.toLowerCase()] || language.toLowerCase();
  }, [language]);

  return (
    <div className="rounded-lg overflow-hidden border border-editor-border bg-editor-surface my-3">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-2 bg-editor-bg/50 border-b border-editor-border">
        <div className="flex items-center gap-2">
          <FileCode className="w-4 h-4 text-editor-muted" />
          {filename ? (
            <span className="text-sm text-editor-text font-mono">{filename}</span>
          ) : (
            <span className="text-sm text-editor-muted">{normalizedLanguage}</span>
          )}
        </div>
        <button
          onClick={handleCopy}
          className="flex items-center gap-1.5 px-2 py-1 text-xs text-editor-muted hover:text-editor-text hover:bg-editor-surface rounded transition-smooth"
        >
          {copied ? (
            <>
              <Check className="w-3.5 h-3.5 text-editor-success" />
              <span className="text-editor-success">Copied!</span>
            </>
          ) : (
            <>
              <Copy className="w-3.5 h-3.5" />
              <span>Copy</span>
            </>
          )}
        </button>
      </div>

      {/* Code */}
      <div className="overflow-x-auto">
        <Highlight theme={themes.nightOwl} code={code.trim()} language={normalizedLanguage}>
          {({ className, style, tokens, getLineProps, getTokenProps }: RenderProps) => (
            <pre
              className={`${className} p-4 text-sm font-mono leading-relaxed`}
              style={{ ...style, background: 'transparent', margin: 0 }}
            >
              {tokens.map((line, i) => {
                const lineProps = getLineProps({ line, key: i });
                const isLastLine = i === tokens.length - 1;

                return (
                  <div
                    key={i}
                    {...lineProps}
                    className={`${lineProps.className || ''} table-row`}
                  >
                    {showLineNumbers && (
                      <span className="table-cell pr-4 text-editor-muted/50 select-none text-right w-8">
                        {i + 1}
                      </span>
                    )}
                    <span className="table-cell">
                      {line.map((token, key) => (
                        <span key={key} {...getTokenProps({ token, key })} />
                      ))}
                      {isStreaming && isLastLine && (
                        <span className="inline-block w-2 h-4 bg-editor-accent animate-pulse ml-0.5 align-middle" />
                      )}
                    </span>
                  </div>
                );
              })}
            </pre>
          )}
        </Highlight>
      </div>
    </div>
  );
};

// Simple inline code component
export const InlineCode: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <code className="px-1.5 py-0.5 rounded bg-editor-surface text-editor-accent text-sm font-mono border border-editor-border">
    {children}
  </code>
);
