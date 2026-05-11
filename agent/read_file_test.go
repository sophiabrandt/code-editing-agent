package agent

import (
	"fmt"
	"testing"
)

func TestReadFileTool(t *testing.T) {
	// Test the ReadFile function
	input := []byte(`{"path":"app/app.go"}`)
	content, err := ReadFile(input)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if len(content) == 0 {
		t.Fatal("expected non-empty content")
	}
	fmt.Printf("Read file successfully, content length: %d\n", len(content))
}
