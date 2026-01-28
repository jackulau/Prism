import { useEffect, useRef, useState } from 'react'
import { Loader2, AlertCircle, Globe, Cloud, Server, ExternalLink, RefreshCw } from 'lucide-react'
import { useSandboxStore, SandboxProvider, DeploymentStatus } from '../../store/sandboxStore'

interface BrowserPreviewProps {
  provider?: SandboxProvider
  deploymentStatus?: DeploymentStatus
  onRefresh?: () => void
}

export function BrowserPreview({ provider: propProvider, deploymentStatus, onRefresh }: BrowserPreviewProps) {
  const iframeRef = useRef<HTMLIFrameElement>(null)
  const {
    previewUrl,
    previewContent,
    isLoading,
    error,
    refreshKey,
    provider: storeProvider,
    currentDeployment,
    refreshPreview
  } = useSandboxStore()
  const [iframeLoaded, setIframeLoaded] = useState(false)

  const provider = propProvider || storeProvider
  const status = deploymentStatus || currentDeployment?.status

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

  const handleRefresh = () => {
    if (onRefresh) {
      onRefresh()
    } else {
      refreshPreview()
    }
  }

  // Provider indicator component
  const ProviderIndicator = () => (
    <div className="absolute top-2 right-2 z-20 flex items-center gap-2">
      {provider === 'vercel' ? (
        <div className="flex items-center gap-1.5 px-2 py-1 bg-black/80 text-white rounded-md text-xs font-medium">
          <Cloud size={12} />
          <span>Vercel</span>
        </div>
      ) : (
        <div className="flex items-center gap-1.5 px-2 py-1 bg-blue-600/80 text-white rounded-md text-xs font-medium">
          <Server size={12} />
          <span>Docker</span>
        </div>
      )}
    </div>
  )

  // Deployment status indicator for Vercel
  const DeploymentStatusIndicator = () => {
    if (provider !== 'vercel' || !status) return null

    const statusConfig: Record<DeploymentStatus, { color: string; label: string; animate?: boolean }> = {
      queued: { color: 'bg-yellow-500', label: 'Queued', animate: true },
      building: { color: 'bg-blue-500', label: 'Building', animate: true },
      ready: { color: 'bg-green-500', label: 'Ready' },
      error: { color: 'bg-red-500', label: 'Error' },
      cancelled: { color: 'bg-gray-500', label: 'Cancelled' },
    }

    const config = statusConfig[status]
    if (!config) return null

    return (
      <div className="absolute top-2 left-2 z-20 flex items-center gap-2">
        <div className="flex items-center gap-1.5 px-2 py-1 bg-editor-surface/90 backdrop-blur rounded-md text-xs">
          <div className={`w-2 h-2 rounded-full ${config.color} ${config.animate ? 'animate-pulse' : ''}`} />
          <span className="text-editor-text">{config.label}</span>
        </div>
      </div>
    )
  }

  // No preview available state
  if (!previewUrl && !previewContent) {
    return (
      <div className="h-full flex flex-col items-center justify-center text-editor-muted relative">
        <ProviderIndicator />
        <Globe size={48} className="mb-4 opacity-50" />
        <p className="text-lg font-medium mb-2">No Preview Available</p>
        <p className="text-sm text-center max-w-md">
          {provider === 'vercel'
            ? 'Deploy your project to Vercel to see the preview here.'
            : 'Run a build or create an HTML file to see the preview here.'
          }
        </p>
        {provider === 'vercel' && (
          <p className="text-xs text-center max-w-md mt-2 text-editor-muted/70">
            Vercel provides fast, globally distributed previews with automatic HTTPS.
          </p>
        )}
      </div>
    )
  }

  // Error state
  if (error) {
    return (
      <div className="h-full flex flex-col items-center justify-center text-editor-error relative">
        <ProviderIndicator />
        <AlertCircle size={48} className="mb-4" />
        <p className="text-lg font-medium mb-2">Preview Error</p>
        <p className="text-sm text-center max-w-md text-editor-muted">{error}</p>
        {currentDeployment?.error && (
          <p className="text-xs text-center max-w-md mt-2 text-editor-muted/70">
            {currentDeployment.error}
          </p>
        )}
      </div>
    )
  }

  // Building/Deploying state for Vercel
  if (provider === 'vercel' && (status === 'queued' || status === 'building')) {
    return (
      <div className="h-full flex flex-col items-center justify-center text-editor-muted relative">
        <ProviderIndicator />
        <DeploymentStatusIndicator />
        <Loader2 size={48} className="animate-spin text-editor-accent mb-4" />
        <p className="text-lg font-medium mb-2">
          {status === 'queued' ? 'Deployment Queued' : 'Building...'}
        </p>
        <p className="text-sm text-center max-w-md">
          {status === 'queued'
            ? 'Your deployment is queued and will start shortly.'
            : 'Vercel is building your project. This usually takes less than a minute.'
          }
        </p>
      </div>
    )
  }

  return (
    <div className="h-full relative bg-white">
      <ProviderIndicator />
      <DeploymentStatusIndicator />

      {/* Action buttons */}
      <div className="absolute top-2 left-1/2 -translate-x-1/2 z-20 flex items-center gap-2">
        <button
          onClick={handleRefresh}
          className="flex items-center gap-1.5 px-2 py-1 bg-editor-surface/90 backdrop-blur rounded-md text-xs hover:bg-editor-hover transition-colors"
          title="Refresh preview"
        >
          <RefreshCw size={12} />
          <span>Refresh</span>
        </button>
        {previewUrl && (
          <a
            href={previewUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-1.5 px-2 py-1 bg-editor-surface/90 backdrop-blur rounded-md text-xs hover:bg-editor-hover transition-colors"
            title="Open in new tab"
          >
            <ExternalLink size={12} />
            <span>Open</span>
          </a>
        )}
      </div>

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
