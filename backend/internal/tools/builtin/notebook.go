package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/sandbox"
)

// Jupyter notebook structures

// Notebook represents a Jupyter notebook
type Notebook struct {
	Cells    []Cell           `json:"cells"`
	Metadata NotebookMetadata `json:"metadata"`
	NBFormat int              `json:"nbformat"`
	NBFormatMinor int         `json:"nbformat_minor"`
}

// Cell represents a notebook cell
type Cell struct {
	CellType       string                 `json:"cell_type"` // "code", "markdown", "raw"
	Source         interface{}            `json:"source"`    // Can be string or []string
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Outputs        []CellOutput           `json:"outputs,omitempty"`
	ExecutionCount *int                   `json:"execution_count,omitempty"`
	ID             string                 `json:"id,omitempty"`
}

// CellOutput represents output from a code cell
type CellOutput struct {
	OutputType string                 `json:"output_type"`
	Text       interface{}            `json:"text,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Name       string                 `json:"name,omitempty"`
	EName      string                 `json:"ename,omitempty"`
	EValue     string                 `json:"evalue,omitempty"`
	Traceback  []string               `json:"traceback,omitempty"`
}

// NotebookMetadata represents notebook metadata
type NotebookMetadata struct {
	KernelSpec   map[string]interface{} `json:"kernelspec,omitempty"`
	LanguageInfo map[string]interface{} `json:"language_info,omitempty"`
}

// CellInfo represents a cell in the API response
type CellInfo struct {
	CellNumber int         `json:"cell_number"`
	CellType   string      `json:"cell_type"`
	Source     string      `json:"source"`
	Outputs    interface{} `json:"outputs,omitempty"`
	ID         string      `json:"id,omitempty"`
}

// NotebookReadTool reads Jupyter notebook files
type NotebookReadTool struct {
	sandbox *sandbox.Service
}

// NewNotebookReadTool creates a new notebook read tool
func NewNotebookReadTool(sandbox *sandbox.Service) *NotebookReadTool {
	return &NotebookReadTool{sandbox: sandbox}
}

func (t *NotebookReadTool) Name() string {
	return "notebook_read"
}

func (t *NotebookReadTool) Description() string {
	return "Reads a Jupyter notebook (.ipynb) file and returns all cells with their outputs. Returns cells with type, source code/text, and outputs."
}

func (t *NotebookReadTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"notebook_path": {
				Type:        "string",
				Description: "The path to the Jupyter notebook file (.ipynb)",
			},
		},
		Required: []string{"notebook_path"},
	}
}

func (t *NotebookReadTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	notebookPath, ok := params["notebook_path"].(string)
	if !ok || notebookPath == "" {
		return nil, fmt.Errorf("notebook_path parameter is required")
	}

	// Validate it's an ipynb file
	if !strings.HasSuffix(strings.ToLower(notebookPath), ".ipynb") {
		return nil, fmt.Errorf("file must be a Jupyter notebook (.ipynb)")
	}

	// Read notebook content
	content, err := t.sandbox.GetFileContent(userID, notebookPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read notebook: %w", err)
	}

	// Parse notebook JSON
	var notebook Notebook
	if err := json.Unmarshal([]byte(content), &notebook); err != nil {
		return nil, fmt.Errorf("failed to parse notebook: %w", err)
	}

	// Convert cells to API format
	cells := make([]CellInfo, len(notebook.Cells))
	for i, cell := range notebook.Cells {
		cells[i] = CellInfo{
			CellNumber: i,
			CellType:   cell.CellType,
			Source:     getCellSource(cell.Source),
			Outputs:    formatOutputs(cell.Outputs),
			ID:         cell.ID,
		}
	}

	return map[string]interface{}{
		"path":     notebookPath,
		"cells":    cells,
		"count":    len(cells),
		"metadata": notebook.Metadata,
	}, nil
}

func (t *NotebookReadTool) RequiresConfirmation() bool {
	return false
}

// getCellSource converts cell source to a single string
func getCellSource(source interface{}) string {
	switch s := source.(type) {
	case string:
		return s
	case []interface{}:
		var lines []string
		for _, line := range s {
			if lineStr, ok := line.(string); ok {
				lines = append(lines, lineStr)
			}
		}
		return strings.Join(lines, "")
	case []string:
		return strings.Join(s, "")
	default:
		return ""
	}
}

// formatOutputs simplifies outputs for display
func formatOutputs(outputs []CellOutput) interface{} {
	if len(outputs) == 0 {
		return nil
	}

	var result []map[string]interface{}
	for _, output := range outputs {
		o := map[string]interface{}{
			"output_type": output.OutputType,
		}

		if output.Text != nil {
			o["text"] = getOutputText(output.Text)
		}
		if output.Data != nil {
			o["data"] = output.Data
		}
		if output.EName != "" {
			o["error"] = fmt.Sprintf("%s: %s", output.EName, output.EValue)
		}

		result = append(result, o)
	}

	return result
}

func getOutputText(text interface{}) string {
	switch t := text.(type) {
	case string:
		return t
	case []interface{}:
		var lines []string
		for _, line := range t {
			if lineStr, ok := line.(string); ok {
				lines = append(lines, lineStr)
			}
		}
		return strings.Join(lines, "")
	case []string:
		return strings.Join(t, "")
	default:
		return ""
	}
}

// NotebookEditTool edits Jupyter notebook cells
type NotebookEditTool struct {
	sandbox     *sandbox.Service
	historyRepo *repository.FileHistoryRepository
}

// NewNotebookEditTool creates a new notebook edit tool
func NewNotebookEditTool(sandbox *sandbox.Service, historyRepo *repository.FileHistoryRepository) *NotebookEditTool {
	return &NotebookEditTool{sandbox: sandbox, historyRepo: historyRepo}
}

func (t *NotebookEditTool) Name() string {
	return "notebook_edit"
}

func (t *NotebookEditTool) Description() string {
	return `Edits a Jupyter notebook cell. Supports replace, insert, and delete operations. Cell numbers are 0-indexed. Use edit_mode='insert' to add a new cell at the specified position.`
}

func (t *NotebookEditTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"notebook_path": {
				Type:        "string",
				Description: "The path to the Jupyter notebook file (.ipynb)",
			},
			"cell_number": {
				Type:        "number",
				Description: "The cell index to edit (0-based). For insert, the new cell is inserted at this position.",
			},
			"new_source": {
				Type:        "string",
				Description: "The new source content for the cell (not required for delete)",
			},
			"edit_mode": {
				Type:        "string",
				Description: "The edit operation: 'replace' (default), 'insert', or 'delete'",
				Enum:        []string{"replace", "insert", "delete"},
			},
			"cell_type": {
				Type:        "string",
				Description: "The cell type for insert operation (default: 'code')",
				Enum:        []string{"code", "markdown", "raw"},
			},
		},
		Required: []string{"notebook_path", "cell_number"},
	}
}

func (t *NotebookEditTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	notebookPath, ok := params["notebook_path"].(string)
	if !ok || notebookPath == "" {
		return nil, fmt.Errorf("notebook_path parameter is required")
	}

	cellNumber, ok := params["cell_number"].(float64)
	if !ok {
		return nil, fmt.Errorf("cell_number parameter is required")
	}
	cellIdx := int(cellNumber)

	editMode := "replace"
	if mode, ok := params["edit_mode"].(string); ok && mode != "" {
		editMode = mode
	}

	newSource, _ := params["new_source"].(string)
	cellType := "code"
	if ct, ok := params["cell_type"].(string); ok && ct != "" {
		cellType = ct
	}

	// Validate edit_mode specific requirements
	if editMode != "delete" && newSource == "" {
		return nil, fmt.Errorf("new_source is required for %s operation", editMode)
	}

	// Read notebook content
	content, err := t.sandbox.GetFileContent(userID, notebookPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read notebook: %w", err)
	}

	// Parse notebook JSON
	var notebook Notebook
	if err := json.Unmarshal([]byte(content), &notebook); err != nil {
		return nil, fmt.Errorf("failed to parse notebook: %w", err)
	}

	// Validate cell index
	switch editMode {
	case "replace", "delete":
		if cellIdx < 0 || cellIdx >= len(notebook.Cells) {
			return nil, fmt.Errorf("cell_number %d is out of range (0-%d)", cellIdx, len(notebook.Cells)-1)
		}
	case "insert":
		if cellIdx < 0 || cellIdx > len(notebook.Cells) {
			return nil, fmt.Errorf("cell_number %d is out of range for insert (0-%d)", cellIdx, len(notebook.Cells))
		}
	}

	// Save to history before modifying
	if t.historyRepo != nil {
		_, _ = t.historyRepo.Create(userID, notebookPath, content, "notebook_edit")
	}

	// Perform the edit
	switch editMode {
	case "replace":
		// Replace cell source, preserve outputs for code cells
		notebook.Cells[cellIdx].Source = strings.Split(newSource, "\n")
		if cellType != "" {
			notebook.Cells[cellIdx].CellType = cellType
		}

	case "insert":
		newCell := Cell{
			CellType: cellType,
			Source:   strings.Split(newSource, "\n"),
			Metadata: make(map[string]interface{}),
		}
		if cellType == "code" {
			newCell.Outputs = []CellOutput{}
		}

		// Insert at position
		notebook.Cells = append(notebook.Cells[:cellIdx], append([]Cell{newCell}, notebook.Cells[cellIdx:]...)...)

	case "delete":
		notebook.Cells = append(notebook.Cells[:cellIdx], notebook.Cells[cellIdx+1:]...)
	}

	// Serialize back to JSON with indentation
	newContent, err := json.MarshalIndent(notebook, "", " ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize notebook: %w", err)
	}

	// Write back
	if err := t.sandbox.WriteFile(userID, notebookPath, string(newContent)); err != nil {
		return nil, fmt.Errorf("failed to write notebook: %w", err)
	}

	return map[string]interface{}{
		"success":     true,
		"path":        notebookPath,
		"cell_number": cellIdx,
		"edit_mode":   editMode,
		"cell_count":  len(notebook.Cells),
	}, nil
}

func (t *NotebookEditTool) RequiresConfirmation() bool {
	return true
}
