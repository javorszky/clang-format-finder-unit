package clang_format

import (
	"os"
	"path/filepath"
)

func deleteContents(dir string) error {
	// Walk through the directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip the root directory itself
		if path == dir {
			return nil
		}

		// Remove the file or directory
		if info.IsDir() {
			// If it's a directory, remove it and all its contents
			err = os.RemoveAll(path)
			if err != nil {
				return err
			}
		} else {
			// If it's a file, just remove it
			err = os.Remove(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
