package agent

import (
	"fmt"
	"testing"
)

func TestListFilesTool(t *testing.T) {
	// Test listing files in the current directory
	input := []byte(`{"path":"."}`)
	content, err := ListFiles(input)
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}
	if len(content) == 0 {
		t.Fatal("expected non-empty content")
	}
	fmt.Printf("Listed files successfully, content: %s\n", content)
}
