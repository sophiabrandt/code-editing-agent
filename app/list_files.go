package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ListFilesInput defines the input for the list_files tool
type ListFilesInput struct {
	Path string `json:"path,omitempty" jsonschema_description:"Optional relative path to list files from. Defaults to current directory if not provided."`
}

// ListFilesDefinition is the tool definition for listing files
var ListFilesDefinition = ToolDefinition{
	Name:        "list_files",
	Description: "List files and directories at a given path. If no path is provided, lists files in the current directory.",
	InputSchema: generateSchema[ListFilesInput](),
	Function:    ListFiles,
}

// ListFiles lists files in a directory recursively using filepath.Walk
func ListFiles(input json.RawMessage) (string, error) {
	var listFilesInput ListFilesInput
	if err := json.Unmarshal(input, &listFilesInput); err != nil {
		return "", fmt.Errorf("failed to parse input: %w", err)
	}

	// Default to current directory if no path provided
	dir := "."
	if listFilesInput.Path != "" {
		dir = listFilesInput.Path
	}

	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// If an error occurs while walking, skip this file/directory and continue
		if err != nil {
			return nil // Skip errors and continue
		}

		// Get relative path from the starting directory
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return nil // Skip if we can't get relative path
		}

		// Skip the base directory itself
		if relPath == "." || relPath == "" {
			return nil
		}

		// Add trailing slash for directories
		if info.IsDir() {
			files = append(files, relPath+"/")
		} else {
			files = append(files, relPath)
		}
		return nil
	})

	// Check if the error is due to the directory not existing
	if err != nil && (os.IsNotExist(err) || errors.Is(err, filepath.SkipDir)) {
		// If the directory doesn't exist, return a user-friendly error
		return "", fmt.Errorf("directory does not exist: %s", dir)
	}
	// If there was another error, we still return it (though we skip errors in walk)
	if err != nil {
		return "", fmt.Errorf("failed to list files: %w", err)
	}

	// Convert results to JSON array
	output, err := json.MarshalIndent(files, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format output: %w", err)
	}

	return string(output), nil
}
