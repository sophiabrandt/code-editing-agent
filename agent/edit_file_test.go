package agent

import (
	"fmt"
	"testing"
)

func TestEditFileTool(t *testing.T) {
	// Test creating a new file
	input := []byte(`{"path":"test.txt","old_str":"","new_str":"Hello World"}`)
	result, err := EditFile(input)
	if err != nil {
		t.Fatalf("EditFile failed: %v", err)
	}
	if result == "" {
		t.Fatal("expected non-empty result")
	}
	fmt.Printf("Create file result: %s\n", result)

	// Test editing an existing file
	input2 := []byte(`{"path":"test.txt","old_str":"Hello","new_str":"Hi"}`)
	result2, err2 := EditFile(input2)
	if err2 != nil {
		t.Fatalf("EditFile failed: %v", err2)
	}
	if result2 == "" {
		t.Fatal("expected non-empty result")
	}
	fmt.Printf("Edit file result: %s\n", result2)

	// Test error case: old_str not found
	input3 := []byte(`{"path":"test.txt","old_str":"Goodbye","new_str":"Farewell"}`)
	result3, err3 := EditFile(input3)
	if result3 != "" || err3 == nil {
		t.Fatal("expected error for old_str not found")
	}
	fmt.Printf("Expected error: %v\n", err3)
}
