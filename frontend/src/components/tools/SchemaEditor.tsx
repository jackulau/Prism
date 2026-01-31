import { useState, useEffect, useCallback } from 'react';
import Editor, { OnChange, OnMount } from '@monaco-editor/react';
import { AlertCircle, CheckCircle } from 'lucide-react';
import type { editor } from 'monaco-editor';

interface SchemaEditorProps {
  value: string;
  onChange: (value: string) => void;
  height?: string;
  readOnly?: boolean;
}

interface ValidationResult {
  valid: boolean;
  errors: string[];
}

function validateJsonSchema(schemaStr: string): ValidationResult {
  if (!schemaStr.trim()) {
    return { valid: true, errors: [] };
  }

  try {
    const schema = JSON.parse(schemaStr);

    const errors: string[] = [];

    if (typeof schema !== 'object' || schema === null) {
      errors.push('Schema must be an object');
      return { valid: false, errors };
    }

    if (schema.type && !['object', 'array', 'string', 'number', 'integer', 'boolean', 'null'].includes(schema.type)) {
      errors.push(`Invalid type: "${schema.type}". Must be one of: object, array, string, number, integer, boolean, null`);
    }

    if (schema.properties && typeof schema.properties !== 'object') {
      errors.push('"properties" must be an object');
    }

    if (schema.required && !Array.isArray(schema.required)) {
      errors.push('"required" must be an array');
    }

    if (schema.properties && schema.required && Array.isArray(schema.required)) {
      const propNames = Object.keys(schema.properties);
      for (const req of schema.required) {
        if (!propNames.includes(req)) {
          errors.push(`Required property "${req}" is not defined in properties`);
        }
      }
    }

    return { valid: errors.length === 0, errors };
  } catch (e) {
    return {
      valid: false,
      errors: [`Invalid JSON: ${e instanceof Error ? e.message : 'Parse error'}`],
    };
  }
}

export function SchemaEditor({ value, onChange, height = '200px', readOnly = false }: SchemaEditorProps) {
  const [validation, setValidation] = useState<ValidationResult>({ valid: true, errors: [] });
  const [editorInstance, setEditorInstance] = useState<editor.IStandaloneCodeEditor | null>(null);

  useEffect(() => {
    const result = validateJsonSchema(value);
    setValidation(result);
  }, [value]);

  const handleEditorDidMount: OnMount = useCallback((editor) => {
    setEditorInstance(editor);
  }, []);

  const handleEditorChange: OnChange = useCallback(
    (newValue) => {
      if (newValue !== undefined) {
        onChange(newValue);
      }
    },
    [onChange]
  );

  const handleFormat = useCallback(() => {
    if (editorInstance && value.trim()) {
      try {
        const formatted = JSON.stringify(JSON.parse(value), null, 2);
        onChange(formatted);
      } catch {
        // Can't format invalid JSON
      }
    }
  }, [editorInstance, value, onChange]);

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {value.trim() && (
            validation.valid ? (
              <div className="flex items-center gap-1 text-xs text-green-400">
                <CheckCircle size={12} />
                <span>Valid JSON Schema</span>
              </div>
            ) : (
              <div className="flex items-center gap-1 text-xs text-red-400">
                <AlertCircle size={12} />
                <span>Invalid</span>
              </div>
            )
          )}
        </div>
        {!readOnly && value.trim() && (
          <button
            type="button"
            onClick={handleFormat}
            className="text-xs text-editor-muted hover:text-editor-text transition-colors"
          >
            Format JSON
          </button>
        )}
      </div>

      <div className="border border-editor-border rounded-lg overflow-hidden">
        <Editor
          height={height}
          language="json"
          value={value}
          onChange={handleEditorChange}
          onMount={handleEditorDidMount}
          theme="vs-dark"
          options={{
            fontSize: 13,
            fontFamily: 'JetBrains Mono, Menlo, Monaco, Consolas, monospace',
            minimap: { enabled: false },
            scrollBeyondLastLine: false,
            wordWrap: 'on',
            lineNumbers: 'on',
            renderLineHighlight: 'line',
            tabSize: 2,
            insertSpaces: true,
            automaticLayout: true,
            padding: { top: 8, bottom: 8 },
            readOnly,
            scrollbar: {
              verticalScrollbarSize: 8,
              horizontalScrollbarSize: 8,
            },
          }}
        />
      </div>

      {validation.errors.length > 0 && (
        <div className="p-2 bg-red-500/10 border border-red-500/30 rounded-lg">
          <ul className="text-xs text-red-400 space-y-1">
            {validation.errors.map((error, index) => (
              <li key={index}>{error}</li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}

export default SchemaEditor;
