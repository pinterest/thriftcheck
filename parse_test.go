// Copyright 2025 Pinterest
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package thriftcheck

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"go.uber.org/thriftrw/ast"
)

const testStructContent = `struct TestStruct { 1: string field }`

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, fmt.Errorf("read error")
}

func TestParseReadError(t *testing.T) {
	_, _, err := Parse(failingReader{})
	if err == nil {
		t.Fatal("expected read error")
	}
	if err.Error() != "read error" {
		t.Errorf("expected 'read error', got: %v", err)
	}
}

func TestFileParserReadError(t *testing.T) {
	parser := NewFileParser(nil)

	absPath, _ := filepath.Abs("test.thrift")
	_, _, err := parser.Parse(failingReader{}, absPath)
	if err == nil {
		t.Fatal("expected read error")
	}
}

func TestFileParserParseFileAbsoluteNotFound(t *testing.T) {
	parser := NewFileParser(nil)

	absPath := "/nonexistent/absolute/path/test.thrift"
	_, _, err := parser.ParseFile(absPath)
	if err == nil {
		t.Fatal("expected error for nonexistent absolute path")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestFileParserCacheHit(t *testing.T) {
	parser := NewFileParser(nil)

	absPath, err := filepath.Abs("test.thrift")
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	prog1, info1, err := parser.Parse(strings.NewReader(testStructContent), absPath)
	if err != nil {
		t.Fatalf("first parse failed: %v", err)
	}

	prog2, info2, err := parser.Parse(strings.NewReader(testStructContent), absPath)
	if err != nil {
		t.Fatalf("second parse failed: %v", err)
	}

	if prog1 != prog2 {
		t.Error("expected cached program (same pointer)")
	}
	if info1 != info2 {
		t.Error("expected cached info (same pointer)")
	}
}

func TestFileParserCacheMiss(t *testing.T) {
	parser := NewFileParser(nil)

	content1 := `struct Struct1 { 1: string field1 }`
	content2 := `struct Struct2 { 1: string field2 }`

	absPath1, _ := filepath.Abs("test1.thrift")
	absPath2, _ := filepath.Abs("test2.thrift")

	prog1, _, err := parser.Parse(strings.NewReader(content1), absPath1)
	if err != nil {
		t.Fatalf("parse 1 failed: %v", err)
	}

	prog2, _, err := parser.Parse(strings.NewReader(content2), absPath2)
	if err != nil {
		t.Fatalf("parse 2 failed: %v", err)
	}

	if prog1 == prog2 {
		t.Error("expected different programs for different files")
	}

	if len(parser.cache) != 2 {
		t.Errorf("expected 2 cache entries, got %d", len(parser.cache))
	}
}

func TestFileParserErrorNotCached(t *testing.T) {
	parser := NewFileParser(nil)

	invalidContent := `struct Invalid {` // Missing closing brace
	absPath, _ := filepath.Abs("invalid.thrift")

	_, _, err1 := parser.Parse(strings.NewReader(invalidContent), absPath)
	if err1 == nil {
		t.Fatal("expected parse error for invalid content")
	}

	if len(parser.cache) != 0 {
		t.Errorf("expected empty cache after error, got %d entries", len(parser.cache))
	}

	_, _, err2 := parser.Parse(strings.NewReader(invalidContent), absPath)
	if err2 == nil {
		t.Fatal("expected parse error on retry")
	}

	if len(parser.cache) != 0 {
		t.Errorf("expected empty cache after retry, got %d entries", len(parser.cache))
	}
}

func TestFileParserAbsolutePathNormalization(t *testing.T) {
	parser := NewFileParser(nil)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	relPath := "test.thrift"
	absPath := filepath.Join(cwd, relPath)

	prog1, _, err := parser.Parse(strings.NewReader(testStructContent), relPath)
	if err != nil {
		t.Fatalf("parse with relative path failed: %v", err)
	}

	prog2, _, err := parser.Parse(strings.NewReader(testStructContent), absPath)
	if err != nil {
		t.Fatalf("parse with absolute path failed: %v", err)
	}

	if prog1 != prog2 {
		t.Error("expected same cached program for relative and absolute paths")
	}

	if len(parser.cache) != 1 {
		t.Errorf("expected 1 cache entry, got %d", len(parser.cache))
	}
}

func TestFileParserConcurrentAccess(t *testing.T) {
	parser := NewFileParser(nil)

	const numGoroutines = 10
	const numIterations = 100

	absPath, _ := filepath.Abs("concurrent.thrift")

	var wg sync.WaitGroup

	// Launch multiple goroutines that all parse the same file concurrently
	for range numGoroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range numIterations {
				_, _, err := parser.Parse(strings.NewReader(testStructContent), absPath)
				if err != nil {
					t.Errorf("concurrent parse error: %v", err)
					return
				}
			}
		}()
	}

	wg.Wait()

	if len(parser.cache) != 1 {
		t.Errorf("expected 1 cache entry, got %d", len(parser.cache))
	}
}

func TestFileParserConcurrentDifferentFiles(t *testing.T) {
	parser := NewFileParser(nil)

	const numGoroutines = 10

	var wg sync.WaitGroup

	// Each goroutine parses a different file
	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			content := fmt.Sprintf("struct TestStruct%d { 1: string field }", id)
			absPath, _ := filepath.Abs(fmt.Sprintf("test%d.thrift", id))

			_, _, err := parser.Parse(strings.NewReader(content), absPath)
			if err != nil {
				t.Errorf("concurrent parse error: %v", err)
			}
		}(i)
	}

	wg.Wait()

	if len(parser.cache) != numGoroutines {
		t.Errorf("expected %d cache entries, got %d", numGoroutines, len(parser.cache))
	}
}

func TestFileParserParseFile(t *testing.T) {
	// Create temporary directory structure for testing
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Create test file in subdirectory
	testFile := filepath.Join(subDir, "test.thrift")
	content := []byte(`struct TestStruct { 1: string field }`)
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Test 1: Parse with absolute path
	parser := NewFileParser(nil)
	_, _, err := parser.ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile with absolute path failed: %v", err)
	}

	// Test 2: Parse with relative path and includes directory
	parser2 := NewFileParser([]string{subDir})
	_, _, err = parser2.ParseFile("test.thrift")
	if err != nil {
		t.Fatalf("ParseFile with relative path failed: %v", err)
	}

	// Test 3: File not found
	parser3 := NewFileParser(nil)
	_, _, err = parser3.ParseFile("nonexistent.thrift")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestFileParserParseFileCurrentDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two subdirectories
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")
	if err := os.Mkdir(dir1, 0755); err != nil {
		t.Fatalf("failed to create dir1: %v", err)
	}
	if err := os.Mkdir(dir2, 0755); err != nil {
		t.Fatalf("failed to create dir2: %v", err)
	}

	// Create same-named file in both directories with different content
	file1 := filepath.Join(dir1, "test.thrift")
	file2 := filepath.Join(dir2, "test.thrift")

	if err := os.WriteFile(file1, []byte(`struct Struct1 { 1: string field1 }`), 0644); err != nil {
		t.Fatalf("failed to write file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte(`struct Struct2 { 1: string field2 }`), 0644); err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}

	// Parser with dir2 in includes, but parsing from dir1
	// Should find test.thrift in dir1 (current directory) not dir2
	parser := NewFileParser([]string{dir2})

	// Parse test.thrift from dir1 - should use dir1 version, not dir2
	prog, _, err := parser.ParseFile(filepath.Join(dir1, "test.thrift"))
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Check that we got Struct1 (from dir1) not Struct2 (from dir2)
	if len(prog.Definitions) == 0 {
		t.Fatal("expected at least one definition")
	}

	structDef, ok := prog.Definitions[0].(*ast.Struct)
	if !ok {
		t.Fatal("expected struct definition")
	}

	if structDef.Name != "Struct1" {
		t.Errorf("expected Struct1 from dir1, got %s", structDef.Name)
	}
}
