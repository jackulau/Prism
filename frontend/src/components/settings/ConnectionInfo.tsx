import { useEffect, useState, useMemo } from 'react';
import { useRemoteAccessStore } from '../../store/remoteAccessStore';
import { useAuthStore } from '../../store/authStore';
import {
  Globe,
  Wifi,
  Copy,
  CheckCircle,
  ExternalLink,
  QrCode,
  X,
} from 'lucide-react';

export function ConnectionInfo() {
  const { accessToken } = useAuthStore();
  const {
    enabled,
    port,
    password,
    publicIP,
    localIPs,
    connectionUrl,
    tlsEnabled,
    setToken,
    fetchConnectionInfo,
  } = useRemoteAccessStore();

  const [copiedField, setCopiedField] = useState<string | null>(null);
  const [showQRModal, setShowQRModal] = useState(false);

  useEffect(() => {
    if (accessToken) {
      setToken(accessToken);
      if (enabled) {
        fetchConnectionInfo();
      }
    }
  }, [accessToken, enabled, setToken, fetchConnectionInfo]);

  const handleCopy = async (value: string, field: string) => {
    await navigator.clipboard.writeText(value);
    setCopiedField(field);
    setTimeout(() => setCopiedField(null), 2000);
  };

  // Generate QR code data URL
  const qrDataUrl = useMemo(() => {
    if (!connectionUrl || !password) return null;

    // Create connection data as JSON
    const connectionData = JSON.stringify({
      url: connectionUrl,
      password: password,
    });

    // Simple QR code generation using a data URL approach
    // In production, you'd use a library like qrcode.react
    // For now, we'll use a Google Charts API fallback
    const encoded = encodeURIComponent(connectionData);
    return `https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encoded}`;
  }, [connectionUrl, password]);

  if (!enabled) {
    return null;
  }

  return (
    <div className="space-y-4">
      <h4 className="text-sm font-medium text-editor-muted">Connection Details</h4>

      {/* Public IP */}
      {publicIP && (
        <div className="flex items-center justify-between p-3 bg-editor-bg rounded-lg">
          <div className="flex items-center gap-3">
            <Globe className="w-4 h-4 text-editor-muted" />
            <div>
              <p className="text-xs text-editor-muted">Public IP</p>
              <p className="font-mono text-sm">{publicIP}</p>
            </div>
          </div>
          <CopyButton
            value={publicIP}
            field="publicIP"
            copiedField={copiedField}
            onCopy={handleCopy}
          />
        </div>
      )}

      {/* Local IPs */}
      {localIPs.length > 0 && (
        <div className="p-3 bg-editor-bg rounded-lg space-y-2">
          <div className="flex items-center gap-2 text-editor-muted">
            <Wifi className="w-4 h-4" />
            <span className="text-xs">Local IP Addresses</span>
          </div>
          {localIPs.map((ip, index) => (
            <div key={ip} className="flex items-center justify-between pl-6">
              <p className="font-mono text-sm">{ip}</p>
              <CopyButton
                value={ip}
                field={`localIP-${index}`}
                copiedField={copiedField}
                onCopy={handleCopy}
              />
            </div>
          ))}
        </div>
      )}

      {/* Port */}
      <div className="flex items-center justify-between p-3 bg-editor-bg rounded-lg">
        <div>
          <p className="text-xs text-editor-muted">Port</p>
          <p className="font-mono text-sm">{port}</p>
        </div>
        <CopyButton
          value={String(port)}
          field="port"
          copiedField={copiedField}
          onCopy={handleCopy}
        />
      </div>

      {/* Full Connection URL */}
      {connectionUrl && (
        <div className="p-3 bg-editor-bg rounded-lg">
          <div className="flex items-center justify-between mb-2">
            <p className="text-xs text-editor-muted">Connection URL</p>
            <div className="flex items-center gap-1">
              {tlsEnabled && (
                <span className="text-xs px-2 py-0.5 bg-green-500/20 text-green-500 rounded">
                  TLS
                </span>
              )}
            </div>
          </div>
          <div className="flex items-center justify-between gap-2">
            <a
              href={connectionUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="font-mono text-sm text-editor-accent hover:underline flex items-center gap-1 truncate"
            >
              {connectionUrl}
              <ExternalLink className="w-3 h-3 flex-shrink-0" />
            </a>
            <CopyButton
              value={connectionUrl}
              field="url"
              copiedField={copiedField}
              onCopy={handleCopy}
            />
          </div>
        </div>
      )}

      {/* QR Code Button */}
      {qrDataUrl && (
        <button
          onClick={() => setShowQRModal(true)}
          className="w-full flex items-center justify-center gap-2 p-3 bg-editor-bg rounded-lg hover:bg-editor-surface transition-colors"
        >
          <QrCode className="w-5 h-5" />
          <span className="text-sm">Show QR Code for Mobile Setup</span>
        </button>
      )}

      {/* QR Code Modal */}
      {showQRModal && qrDataUrl && (
        <>
          <div
            className="fixed inset-0 bg-black/50 z-40"
            onClick={() => setShowQRModal(false)}
          />
          <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[320px] max-w-[90vw] bg-editor-bg border border-editor-border rounded-lg shadow-xl z-50 p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold">Scan to Connect</h3>
              <button
                onClick={() => setShowQRModal(false)}
                className="p-1 hover:bg-editor-surface rounded"
              >
                <X className="w-5 h-5 text-editor-muted" />
              </button>
            </div>

            <div className="flex flex-col items-center gap-4">
              <div className="bg-white p-4 rounded-lg">
                <img
                  src={qrDataUrl}
                  alt="Connection QR Code"
                  className="w-48 h-48"
                />
              </div>
              <p className="text-sm text-editor-muted text-center">
                Scan this QR code with your mobile device to quickly connect to this instance.
              </p>
            </div>
          </div>
        </>
      )}
    </div>
  );
}

function CopyButton({
  value,
  field,
  copiedField,
  onCopy,
}: {
  value: string;
  field: string;
  copiedField: string | null;
  onCopy: (value: string, field: string) => void;
}) {
  const isCopied = copiedField === field;

  return (
    <button
      onClick={() => onCopy(value, field)}
      className="p-1.5 hover:bg-editor-surface rounded text-editor-muted hover:text-editor-text transition-colors"
      title="Copy"
    >
      {isCopied ? (
        <CheckCircle className="w-4 h-4 text-green-500" />
      ) : (
        <Copy className="w-4 h-4" />
      )}
    </button>
  );
}
