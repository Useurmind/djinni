package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/useurmind/djinni/pkg/config"
	"github.com/useurmind/djinni/pkg/log"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean all djinni storage",
	Long:  `Remove all contents from the configured storage base directory`,
	RunE:  runClean,
}

func runClean(cmd *cobra.Command, args []string) error {
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return fmt.Errorf("failed to get force flag: %w", err)
	}

	globalCfg, err := config.LoadGlobalConfig()
	if err != nil {
		return fmt.Errorf("failed to load global config: %w", err)
	}

	baseDir := globalCfg.StorageBaseDirectory

	dirsToClean, err := findStorageDirs(baseDir)
	if err != nil {
		return fmt.Errorf("failed to find storage directories: %w", err)
	}

	if len(dirsToClean) == 0 {
		log.Info("No storage directories found to clean")
		return nil
	}

	if !force {
		if !confirmCleanup(dirsToClean) {
			log.Info("Cleanup cancelled")
			return nil
		}
	}

	if err := cleanStorageDirs(dirsToClean); err != nil {
		return fmt.Errorf("failed to clean storage: %w", err)
	}

	log.Success("Storage cleaned successfully")
	return nil
}

func findStorageDirs(baseDir string) ([]string, error) {
	var dirs []string

	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return dirs, nil
	}

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read base directory %s: %w", baseDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		repoPath := filepath.Join(baseDir, entry.Name())
		dirs = append(dirs, repoPath)

		repoEntries, err := os.ReadDir(repoPath)
		if err != nil {
			log.Error(fmt.Sprintf("Failed to read repository directory %s: %v", repoPath, err))
			continue
		}

		for _, repoEntry := range repoEntries {
			if !repoEntry.IsDir() {
				continue
			}

			agentPath := filepath.Join(repoPath, repoEntry.Name())
			dirs = append(dirs, agentPath)

			agentEntries, err := os.ReadDir(agentPath)
			if err != nil {
				log.Error(fmt.Sprintf("Failed to read agent directory %s: %v", agentPath, err))
				continue
			}

			for _, agentEntry := range agentEntries {
				if !agentEntry.IsDir() {
					continue
				}

				path := filepath.Join(agentPath, agentEntry.Name())
				dirs = append(dirs, path)
			}
		}
	}

	return dirs, nil
}

func confirmCleanup(dirs []string) bool {
	fmt.Println("The following directories will be deleted:")
	fmt.Println()

	for _, dir := range dirs {
		fmt.Printf("  - %s\n", dir)
	}

	fmt.Println()
	fmt.Print("Are you sure you want to delete all these directories? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(response)
	return strings.EqualFold(response, "y") || strings.EqualFold(response, "yes")
}

func cleanStorageDirs(dirs []string) error {
	for _, dir := range dirs {
		if err := removeDir(dir); err != nil {
			return fmt.Errorf("failed to remove %s: %w", dir, err)
		}
		log.Info(fmt.Sprintf("Deleted: %s", dir))
	}

	return nil
}

func removeDir(path string) error {
	cmd := exec.Command("podman", "unshare", "rm", "-rf", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove %s: %w, output: %s", path, err, strings.TrimSpace(string(output)))
	}
	return nil
}

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}
