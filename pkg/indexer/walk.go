package indexer

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

func (this *Indexer) Walk(ctx context.Context) error {
	err := filepath.WalkDir(this.path, func(path string, file fs.DirEntry, err error) error {
		// Propogate errors if they are cancel/timeout
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		// Ignore hidden files and folders
		if strings.HasPrefix(file.Name(), ".") {
			if file.IsDir() {
				return filepath.SkipDir
			}
			return err
		}
		// Process file, but ignore any errors
		if info, err := file.Info(); err == nil {
			if err := this.ProcessPath(EVENT_TYPE_ADDED, path, info); err != nil {
				fmt.Println("Error:", err)
			}
		}
		// Return any context error
		return ctx.Err()
	})
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil
	} else {
		return err
	}
}
