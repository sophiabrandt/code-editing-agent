package agent

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/invopop/jsonschema"
)

// ReadFileInput defines the input for the read_file tool
type ReadFileInput struct {
	Path string `json:"path" jsonschema_description:"The relative path of a file in the working directory."`
}

// ReadFile is the function that reads a file and returns its contents
func ReadFile(input json.RawMessage) (string, error) {
	var fileInput ReadFileInput
	if err := json.Unmarshal(input, &fileInput); err != nil {
		return "", fmt.Errorf("failed to parse input: %w", err)
	}

	content, err := os.ReadFile(fileInput.Path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %q: %w", fileInput.Path, err)
	}
	return string(content), nil
}

// generateSchema creates a JSON schema for a given type using jsonschema
func generateSchema[T any]() json.RawMessage {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	bytes, _ := json.Marshal(schema)
	return bytes
}

// ReadFileDefinition is the tool definition for reading files
var ReadFileDefinition = ToolDefinition{
	Name:        "read_file",
	Description: "Read the contents of a given relative file path. Use this when you want to see what's inside a file. Do not use this with directory names.",
	InputSchema: generateSchema[ReadFileInput](),
	Function:    ReadFile,
}
