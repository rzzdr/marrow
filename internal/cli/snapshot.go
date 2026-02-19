package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rzzdr/marrow/internal/model"
	"github.com/spf13/cobra"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage context snapshots",
}

var snapshotName string

var snapshotCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a snapshot of the current .marrow/ state",
	RunE: func(cmd *cobra.Command, args []string) error {
		s := getStoreFromRoot()
		if !s.Exists() {
			return fmt.Errorf("no .marrow/ found")
		}

		src := s.Root()
		fullName := time.Now().UTC().Format("20060102T150405") + "_" + snapshotName
		dst := filepath.Join(src, "snapshots", fullName)
		if _, err := os.Stat(dst); err == nil {
			return fmt.Errorf("snapshot %q already exists", fullName)
		}

		if err := copyDir(src, dst); err != nil {
			return fmt.Errorf("creating snapshot: %w", err)
		}

		_ = s.AppendChangelog(model.ChangelogEntry{
			Action:  "snapshot_created",
			Summary: fullName,
		})

		fmt.Printf("Snapshot created: %s\n", fullName)
		return nil
	},
}

var snapshotListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available snapshots",
	RunE: func(cmd *cobra.Command, args []string) error {
		s := getStoreFromRoot()
		dir := filepath.Join(s.Root(), "snapshots")

		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No snapshots.")
				return nil
			}
			return err
		}

		for _, e := range entries {
			if e.IsDir() {
				fmt.Println("  " + e.Name())
			}
		}
		return nil
	},
}

func init() {
	snapshotCreateCmd.Flags().StringVar(&snapshotName, "name", "", "Snapshot name (required)")
	_ = snapshotCreateCmd.MarkFlagRequired("name")

	snapshotCmd.AddCommand(snapshotCreateCmd)
	snapshotCmd.AddCommand(snapshotListCmd)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel(src, path)
		if rel == "snapshots" || strings.HasPrefix(rel, "snapshots"+string(filepath.Separator)) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}

	if _, err = io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
