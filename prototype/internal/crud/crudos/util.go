package crudos

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sis/internal/pk"
)

func (c CrudOs) init() error {
	err := os.MkdirAll(c.root, c.perm)
	if err != nil {
		return fmt.Errorf("error creating root directory: %w", err)
	}
	return nil
}

func (c CrudOs) pkToPath(pk []string) string {
	rootedPk := append([]string{c.root}, pk...)
	return filepath.Join(rootedPk...)
}

func (c CrudOs) absPathToPk(absPath string) (key []string) {
	rootedPk := pk.New(absPath)
	rootPk := pk.New(c.root)
	return rootedPk[len(rootPk):]
}

func (c CrudOs) isDirEmpty(dirPath string) (bool, error) {
	f, err := os.Open(dirPath)
	if err != nil {
		return false, fmt.Errorf("error on directory open: %w", err)
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil // Directory is empty
	}
	if err != nil {
		return false, fmt.Errorf("error reading directory entries: %w", err)
	}
	return false, nil
}
