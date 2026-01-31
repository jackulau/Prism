import { useState, useMemo, useEffect, useRef } from 'react';
import { useWorkflowExecutionStore } from '../../store/workflowExecutionStore';

interface StateInspectorProps {
  className?: string;
}

type JsonValue = string | number | boolean | null | JsonValue[] | { [key: string]: JsonValue };

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      console.error('Failed to copy to clipboard');
    }
  };

  return (
    <button
      onClick={(e) => {
        e.stopPropagation();
        handleCopy();
      }}
      className="p-0.5 rounded hover:bg-gray-700 text-gray-500 hover:text-gray-300 transition-colors opacity-0 group-hover:opacity-100"
      title={copied ? 'Copied!' : 'Copy value'}
    >
      {copied ? (
        <svg className="w-3 h-3 text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
        </svg>
      ) : (
        <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
        </svg>
      )}
    </button>
  );
}

function TreeNode({
  keyName,
  value,
  depth = 0,
  changedKeys,
  searchTerm,
  path = '',
}: {
  keyName: string;
  value: JsonValue;
  depth?: number;
  changedKeys: Set<string>;
  searchTerm: string;
  path?: string;
}) {
  const [isExpanded, setIsExpanded] = useState(depth < 2);
  const currentPath = path ? `${path}.${keyName}` : keyName;
  const isChanged = changedKeys.has(currentPath);
  const matchesSearch = searchTerm && currentPath.toLowerCase().includes(searchTerm.toLowerCase());

  const valueType = typeof value;
  const isObject = value !== null && valueType === 'object';
  const isArray = Array.isArray(value);
  const isEmpty = isObject && Object.keys(value as object).length === 0;

  const getValueColor = () => {
    if (value === null) return 'text-gray-500';
    switch (valueType) {
      case 'string': return 'text-green-400';
      case 'number': return 'text-blue-400';
      case 'boolean': return value ? 'text-yellow-400' : 'text-orange-400';
      default: return 'text-gray-300';
    }
  };

  const formatValue = () => {
    if (value === null) return 'null';
    if (valueType === 'string') return `"${value}"`;
    if (valueType === 'boolean') return value ? 'true' : 'false';
    return String(value);
  };

  const getPreview = () => {
    if (!isObject) return null;
    const obj = value as object;
    const keys = Object.keys(obj);
    if (isArray) {
      return `[${keys.length}]`;
    }
    return `{${keys.length}}`;
  };

  const renderChildren = () => {
    if (!isObject || !isExpanded) return null;
    const obj = value as Record<string, JsonValue>;
    const entries = Object.entries(obj);

    return (
      <div className="ml-4 border-l border-gray-800">
        {entries.map(([key, val]) => (
          <TreeNode
            key={key}
            keyName={isArray ? `[${key}]` : key}
            value={val}
            depth={depth + 1}
            changedKeys={changedKeys}
            searchTerm={searchTerm}
            path={currentPath}
          />
        ))}
      </div>
    );
  };

  return (
    <div>
      <div
        className={`
          group flex items-center gap-1 py-0.5 px-1 hover:bg-gray-800/50 rounded cursor-pointer text-xs
          ${isChanged ? 'bg-yellow-500/10' : ''}
          ${matchesSearch ? 'bg-blue-500/10' : ''}
        `}
        onClick={() => isObject && !isEmpty && setIsExpanded(!isExpanded)}
        style={{ paddingLeft: `${depth * 8}px` }}
      >
        {/* Expand/collapse arrow */}
        {isObject && !isEmpty ? (
          <svg
            className={`w-3 h-3 text-gray-500 transition-transform ${isExpanded ? 'rotate-90' : ''}`}
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
          </svg>
        ) : (
          <span className="w-3" />
        )}

        {/* Key name */}
        <span className={`text-purple-400 ${isChanged ? 'font-semibold' : ''}`}>
          {keyName}
        </span>
        <span className="text-gray-600">:</span>

        {/* Value or preview */}
        {isObject ? (
          <span className="text-gray-500">
            {getPreview()}
            {isEmpty && <span className="ml-1">{isArray ? '[]' : '{}'}</span>}
          </span>
        ) : (
          <span className={`${getValueColor()} truncate max-w-xs`}>
            {formatValue()}
          </span>
        )}

        {/* Changed indicator */}
        {isChanged && (
          <span className="ml-1 w-1.5 h-1.5 rounded-full bg-yellow-400 animate-pulse" />
        )}

        {/* Copy button */}
        <CopyButton text={JSON.stringify(value, null, 2)} />
      </div>

      {renderChildren()}
    </div>
  );
}

export function StateInspector({ className = '' }: StateInspectorProps) {
  const { workflowState } = useWorkflowExecutionStore();
  const [searchTerm, setSearchTerm] = useState('');
  const [changedKeys, setChangedKeys] = useState<Set<string>>(new Set());
  const previousStateRef = useRef<Record<string, unknown>>({});

  // Track changed keys
  useEffect(() => {
    const findChangedKeys = (
      prev: Record<string, unknown>,
      next: Record<string, unknown>,
      path = ''
    ): string[] => {
      const changes: string[] = [];

      for (const key of Object.keys(next)) {
        const currentPath = path ? `${path}.${key}` : key;
        const prevValue = prev[key];
        const nextValue = next[key];

        if (prevValue !== nextValue) {
          changes.push(currentPath);

          // Recursively check nested objects
          if (
            typeof nextValue === 'object' &&
            nextValue !== null &&
            typeof prevValue === 'object' &&
            prevValue !== null
          ) {
            changes.push(
              ...findChangedKeys(
                prevValue as Record<string, unknown>,
                nextValue as Record<string, unknown>,
                currentPath
              )
            );
          }
        }
      }

      return changes;
    };

    const newChanges = findChangedKeys(previousStateRef.current, workflowState);
    if (newChanges.length > 0) {
      setChangedKeys(new Set(newChanges));

      // Clear highlights after 2 seconds
      const timeout = setTimeout(() => {
        setChangedKeys(new Set());
      }, 2000);

      return () => clearTimeout(timeout);
    }

    previousStateRef.current = { ...workflowState };
  }, [workflowState]);

  const filteredState = useMemo(() => {
    if (!searchTerm) return workflowState;

    const filterObject = (obj: Record<string, unknown>, term: string): Record<string, unknown> => {
      const result: Record<string, unknown> = {};

      for (const [key, value] of Object.entries(obj)) {
        if (key.toLowerCase().includes(term.toLowerCase())) {
          result[key] = value;
        } else if (typeof value === 'object' && value !== null && !Array.isArray(value)) {
          const filtered = filterObject(value as Record<string, unknown>, term);
          if (Object.keys(filtered).length > 0) {
            result[key] = filtered;
          }
        }
      }

      return result;
    };

    return filterObject(workflowState, searchTerm);
  }, [workflowState, searchTerm]);

  const stateString = JSON.stringify(workflowState, null, 2);
  const isEmpty = Object.keys(workflowState).length === 0;

  return (
    <div className={`flex flex-col h-full bg-gray-900 ${className}`}>
      {/* Header */}
      <div className="flex-shrink-0 p-3 border-b border-gray-800">
        <div className="flex items-center justify-between mb-2">
          <h3 className="text-sm font-semibold text-gray-200">Workflow State</h3>
          <CopyButton text={stateString} />
        </div>

        {/* Search input */}
        <div className="relative">
          <svg
            className="absolute left-2 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            type="text"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            placeholder="Filter keys..."
            className="w-full pl-8 pr-3 py-1.5 text-xs bg-gray-800 border border-gray-700 rounded focus:outline-none focus:border-blue-500 text-gray-300 placeholder-gray-500"
          />
          {searchTerm && (
            <button
              onClick={() => setSearchTerm('')}
              className="absolute right-2 top-1/2 -translate-y-1/2 p-0.5 rounded hover:bg-gray-700"
            >
              <svg className="w-3 h-3 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          )}
        </div>
      </div>

      {/* State tree */}
      <div className="flex-1 overflow-auto p-2">
        {isEmpty ? (
          <div className="text-center text-gray-500 text-sm py-8">
            No state data
          </div>
        ) : (
          <div className="font-mono">
            {Object.entries(filteredState).map(([key, value]) => (
              <TreeNode
                key={key}
                keyName={key}
                value={value as JsonValue}
                changedKeys={changedKeys}
                searchTerm={searchTerm}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

export default StateInspector;
