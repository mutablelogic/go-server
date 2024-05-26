package nginx

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed conf/*
var content embed.FS

// fsCopyTo copies the content of the embedded filesystem to the destination
func fsCopyTo(dst string) error {
	// Open the source directory
	src, err := content.ReadDir("conf")
	if err != nil {
		return err
	}
	return fsCopyDir(dst, "conf", src)
}

// fsCopyDir copies the content of a directory to the destination
func fsCopyDir(dst string, srcPath string, entries []fs.DirEntry) error {
	for _, entry := range entries {
		srcEntryPath := filepath.Join(srcPath, entry.Name())
		dstEntryPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Create the destination directory
			if err := os.MkdirAll(dstEntryPath, os.ModePerm); err != nil {
				return err
			}
			// Read the directory contents
			subEntries, err := content.ReadDir(srcEntryPath)
			if err != nil {
				return err
			}
			// Recursively copy the directory
			if err := fsCopyDir(dstEntryPath, srcEntryPath, subEntries); err != nil {
				return err
			}
		} else if err := fsCopyFile(dstEntryPath, srcEntryPath); err != nil {
			return err
		}
	}
	return nil
}

// fsCopyFile copies a file from the embedded filesystem to the destination
func fsCopyFile(dstFile string, srcFile string) error {
	fmt.Println("Copying", srcFile, "to", dstFile)
	src, err := content.Open(srcFile)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	// Return success
	return nil
}
