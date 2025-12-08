import { useEffect, useRef, useState } from 'react'
import { Loader2, AlertCircle, Globe } from 'lucide-react'
import { useSandboxStore } from '../../store/sandboxStore'

export function BrowserPreview() {
  const iframeRef = useRef<HTMLIFrameElement>(null)
  const { previewUrl, previewContent, isLoading, error, refreshKey } = useSandboxStore()
  const [iframeLoaded, setIframeLoaded] = useState(false)

  useEffect(() => {
    setIframeLoaded(false)
  }, [refreshKey, previewUrl])

  // Handle inline content preview (for HTML/CSS/JS that doesn't need a server)
  useEffect(() => {
    if (previewContent && iframeRef.current) {
      const iframe = iframeRef.current
      const doc = iframe.contentDocument || iframe.contentWindow?.document
      if (doc) {
        doc.open()
        doc.write(previewContent)
        doc.close()
        setIframeLoaded(true)
      }
    }
  }, [previewContent, refreshKey])

  const handleIframeLoad = () => {
    setIframeLoaded(true)
  }

  // No preview available state
  if (!previewUrl && !previewContent) {
    return (
      <div className="h-full flex flex-col items-center justify-center text-editor-muted">
        <Globe size={48} className="mb-4 opacity-50" />
        <p className="text-lg font-medium mb-2">No Preview Available</p>
        <p className="text-sm text-center max-w-md">
          Run a build or create an HTML file to see the preview here.
          The sandbox will display your running application.
        </p>
      </div>
    )
  }

  // Error state
  if (error) {
    return (
      <div className="h-full flex flex-col items-center justify-center text-editor-error">
        <AlertCircle size={48} className="mb-4" />
        <p className="text-lg font-medium mb-2">Preview Error</p>
        <p className="text-sm text-center max-w-md text-editor-muted">{error}</p>
      </div>
    )
  }

  return (
    <div className="h-full relative bg-white">
      {/* Loading overlay */}
      {(isLoading || !iframeLoaded) && (
        <div className="absolute inset-0 flex items-center justify-center bg-editor-surface z-10">
          <div className="flex flex-col items-center gap-3">
            <Loader2 size={32} className="animate-spin text-editor-accent" />
            <span className="text-sm text-editor-muted">Loading preview...</span>
          </div>
        </div>
      )}

      {/* Preview iframe */}
      {previewUrl ? (
        <iframe
          ref={iframeRef}
          key={refreshKey}
          src={previewUrl}
          className="w-full h-full border-0"
          title="Sandbox Preview"
          sandbox="allow-scripts allow-same-origin allow-forms allow-modals allow-popups"
          onLoad={handleIframeLoad}
        />
      ) : previewContent ? (
        <iframe
          ref={iframeRef}
          key={refreshKey}
          className="w-full h-full border-0"
          title="Sandbox Preview"
          sandbox="allow-scripts allow-same-origin allow-forms allow-modals"
        />
      ) : null}
    </div>
  )
}
