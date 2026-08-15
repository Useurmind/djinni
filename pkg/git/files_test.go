package git

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetChangedFilesWithDiffs_ParseModifiedFilesWithDiffFormat(t *testing.T) {
	tmpDir := t.TempDir()
	binDir := tmpDir
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("Failed to create bin dir: %v", err)
	}

	gitPath := filepath.Join(binDir, "git")
	gitScript := `#!/bin/bash
if [[ "$*" =~ "status" ]] && [[ "$*" =~ "porcelain" ]]; then
    echo "$MOCK_STATUS"
    exit 0
fi
if [[ "$*" =~ "diff" ]] && [[ "$*" =~ "HEAD" ]]; then
    echo "diff --git a/file1.go b/file1.go"
    echo "index 123..456 100644"
    echo "--- a/file1.go"
    echo "+++ b/file1.go"
    echo "@@ -1 +1 @@"
    echo "-old"
    echo "+new"
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
		mockDiff        string
		wantErr         bool
		containsPattern string
	}{
		{
			name:            "parse modified files with diff format",
			repoDir:         tmpDir,
			mockStatus:      " M file1.go\nM  file2.go\nA  file3.go\n",
			wantErr:         false,
			containsPattern: "- file1.go modified",
		},
		{
			name:            "parse untracked files",
			repoDir:         tmpDir,
			mockStatus:      "?? newfile.go\n",
			wantErr:         false,
			containsPattern: "- newfile.go untracked",
		},
		{
			name:            "parse all status types",
			repoDir:         tmpDir,
			mockStatus:      "M  modified.go\nA  added.go\nD  deleted.go\nR  renamed.go\nC  copied.go\nU  unmerged.go\n?? untracked.go\n",
			wantErr:         false,
			containsPattern: "- modified.go modified",
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

			result, err := GetChangedFilesWithDiffs(tt.repoDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetChangedFilesWithDiffs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.containsPattern != "" {
				if !strings.Contains(result, tt.containsPattern) {
					t.Errorf("GetChangedFilesWithDiffs() result = %s, should contain pattern '%s'", result, tt.containsPattern)
				}
			}
		})
	}
}

func TestGetChangedFilesWithDiffs_Name(t *testing.T) {
	result, err := GetChangedFilesWithDiffs("/tmp")
	if err != nil {
		t.Logf("Expected error for non-git directory: %v", err)
	}
	t.Logf("Result for /tmp (no git repo): %s", result)
}
