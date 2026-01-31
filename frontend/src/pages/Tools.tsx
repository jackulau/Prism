import { useState } from 'react'
import { Wrench, History } from 'lucide-react'
import { ToolExecutionLogs } from '../components/tools/ToolExecutionLogs'

type TabType = 'logs'

interface Tab {
  id: TabType
  label: string
  icon: typeof Wrench
}

const tabs: Tab[] = [
  { id: 'logs', label: 'Execution Logs', icon: History },
]

export default function Tools() {
  const [activeTab, setActiveTab] = useState<TabType>('logs')

  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-6xl mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="space-y-2">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-orange-500/20 text-orange-400 rounded-lg">
              <Wrench size={24} />
            </div>
            <div>
              <h1 className="text-2xl font-bold text-editor-text">Tools</h1>
              <p className="text-editor-muted">
                View tool execution history and debug tool runs
              </p>
            </div>
          </div>
        </div>

        {/* Tabs */}
        <div className="border-b border-editor-border">
          <nav className="-mb-px flex gap-4">
            {tabs.map((tab) => {
              const Icon = tab.icon
              return (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`flex items-center gap-2 px-1 py-3 text-sm font-medium border-b-2 transition-colors ${
                    activeTab === tab.id
                      ? 'border-editor-accent text-editor-accent'
                      : 'border-transparent text-editor-muted hover:text-editor-text hover:border-editor-border'
                  }`}
                >
                  <Icon size={16} />
                  {tab.label}
                </button>
              )
            })}
          </nav>
        </div>

        {/* Tab content */}
        <div>
          {activeTab === 'logs' && <ToolExecutionLogs />}
        </div>
      </div>
    </div>
  )
}
