package ai

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestGitChangedFilesTool_Call(t *testing.T) {
	tmpDir := t.TempDir()
	binDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("Failed to create bin dir: %v", err)
	}

	gitPath := filepath.Join(binDir, "git")
	gitScript := `#!/bin/bash
if [[ "$*" =~ "status" ]] && [[ "$*" =~ "porcelain" ]]; then
    echo "$MOCK_STATUS"
    exit 0
fi
exit 128
`
	if err := os.WriteFile(gitPath, []byte(gitScript), 0755); err != nil {
		t.Fatalf("Failed to create git stub: %v", err)
	}

	oldPath := os.Getenv("PATH")
	_ = os.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath)
	t.Cleanup(func() {
		_ = os.Setenv("PATH", oldPath)
	})

	tests := []struct {
		name            string
		repoDir         string
		mockStatus      string
		wantErr         bool
		containsPattern string
	}{
		{
			name:            "parse modified files",
			repoDir:         tmpDir,
			mockStatus:      " M file1.go\nM  file2.go\nA  file3.go\n",
			wantErr:         false,
			containsPattern: "modified",
		},
		{
			name:            "parse untracked files",
			repoDir:         tmpDir,
			mockStatus:      "?? newfile.go\n",
			wantErr:         false,
			containsPattern: "untracked",
		},
		{
			name:            "parse all status types",
			repoDir:         tmpDir,
			mockStatus:      "M  modified.go\nA  added.go\nD  deleted.go\nR  renamed.go\nC  copied.go\nU  unmerged.go\n?? untracked.go\n",
			wantErr:         false,
			containsPattern: "modified",
		},
		{
			name:            "empty status",
			repoDir:         tmpDir,
			mockStatus:      "",
			wantErr:         false,
			containsPattern: "No changes detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv("MOCK_STATUS", tt.mockStatus)

			tool := NewGitChangedFilesTool(tt.repoDir)

			result, err := tool.Call(context.Background(), "")
			if (err != nil) != tt.wantErr {
				t.Errorf("GitChangedFilesTool.Call() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.containsPattern != "" {
				if !contains(result, tt.containsPattern) {
					t.Errorf("GitChangedFilesTool.Call() result = %s, should contain pattern '%s'", result, tt.containsPattern)
				}
			}
		})
	}
}

func TestGitChangedFilesTool_Name(t *testing.T) {
	tool := NewGitChangedFilesTool("/tmp")
	if tool.Name() != "git_changed_files" {
		t.Errorf("Expected name 'git_changed_files', got '%s'", tool.Name())
	}
}

func TestGitChangedFilesTool_Description(t *testing.T) {
	tool := NewGitChangedFilesTool("/tmp")
	desc := tool.Description()
	if len(desc) == 0 {
		t.Error("Description should not be empty")
	}
}

func TestReadFileTool_Call(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		readPaths   []string
		input       string
		wantErr     bool
		containsTxt string
	}{
		{
			name:        "file exists in read path",
			readPaths:   []string{tmpDir},
			input:       "test.txt",
			wantErr:     false,
			containsTxt: "test content",
		},
		{
			name:      "empty input",
			readPaths: []string{tmpDir},
			input:     "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input != "" {
				if err := os.WriteFile(filepath.Join(tmpDir, tt.input), []byte("test content"), 0644); err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
			}

			tool := NewReadFileTool(tt.readPaths)

			result, err := tool.Call(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadFileTool.Call() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.containsTxt != "" {
				if !contains(result, tt.containsTxt) {
					t.Errorf("ReadFileTool.Call() result = %s, should contain '%s'", result, tt.containsTxt)
				}
			}
		})
	}
}

func TestReadFileTool_Name(t *testing.T) {
	tool := NewReadFileTool([]string{"/tmp"})
	if tool.Name() != "read_file" {
		t.Errorf("Expected name 'read_file', got '%s'", tool.Name())
	}
}

func TestReadFileTool_Description(t *testing.T) {
	tool := NewReadFileTool([]string{"/tmp"})
	desc := tool.Description()
	if len(desc) == 0 {
		t.Error("Description should not be empty")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
