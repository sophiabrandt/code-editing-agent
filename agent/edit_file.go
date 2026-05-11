package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EditFileInput defines the input for the edit_file tool
type EditFileInput struct {
	Path   string `json:"path" jsonschema_description:"The path to the file"`
	OldStr string `json:"old_str" jsonschema_description:"Text to search for - must match exactly and must only have one match exactly"`
	NewStr string `json:"new_str" jsonschema_description:"Text to replace old_str with"`
}

// EditFileDefinition is the tool definition for editing files
var EditFileDefinition = ToolDefinition{
	Name:        "edit_file",
	Description: `Make edits to a text file. Replaces 'old_str' with 'new_str' in the given file. 'old_str' and 'new_str' MUST be different from each other. If the file specified with path doesn't exist, it will be created.`,
	InputSchema: generateSchema[EditFileInput](),
	Function:    EditFile,
}

// EditFile edits a file by replacing oldStr with newStr
func EditFile(input json.RawMessage) (string, error) {
	var editInput EditFileInput
	if err := json.Unmarshal(input, &editInput); err != nil {
		return "", fmt.Errorf("failed to parse input: %w", err)
	}

	// Validate input
	if editInput.Path == "" {
		return "", fmt.Errorf("path is required")
	}
	if editInput.OldStr == editInput.NewStr {
		return "", fmt.Errorf("old_str and new_str must be different")
	}

	_, err := os.Stat(editInput.Path)
	if err != nil {
		if os.IsNotExist(err) {
			if editInput.OldStr == "" {
				return createNewFile(editInput.Path, editInput.NewStr)
			}
			return "", fmt.Errorf("file does not exist and old_str is not empty - cannot create file with replacement")
		}
		return "", fmt.Errorf("failed to access file: %w", err)
	}

	content, err := os.ReadFile(editInput.Path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	oldContent := string(content)
	newContent := strings.Replace(oldContent, editInput.OldStr, editInput.NewStr, -1)

	if oldContent == newContent {
		if editInput.OldStr == "" {
			// No change needed - file already matches new content
			return "File already contains the new content", nil
		}
		return "", fmt.Errorf("old_str not found in file")
	}

	if err := os.WriteFile(editInput.Path, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return "OK", nil
}

// createNewFile creates a new file with the given content
func createNewFile(filePath, content string) (string, error) {
	dir := filepath.Dir(filePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("failed to create directory: %w", err)
		}
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}

	return fmt.Sprintf("Successfully created file %s", filePath), nil
}
