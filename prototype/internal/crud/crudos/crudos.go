package crudos

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sis/internal/metrics"
)

type CrudOs struct {
	root string
	perm os.FileMode
}

func New(rootPath string, permissions ...os.FileMode) (CrudOs, error) {
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return CrudOs{}, err
	}
	var currPermissions os.FileMode = 0777
	if len(permissions) > 0 {
		currPermissions = permissions[0]
	}
	c := CrudOs{
		root: absRoot,
		perm: currPermissions,
	}
	err = c.init()
	if err != nil {
		return c, fmt.Errorf("error initializing CrudOs instance: %s", err)
	}
	return c, nil
}

func (c CrudOs) Create(pk []string, blob []byte) error {

	if len(pk) == 0 {
		return fmt.Errorf("pk cannot be empty")
	}

	pkPath := c.pkToPath(pk)
	directoriesPath := c.pkToPath(pk[:len(pk)-1])
	err := os.MkdirAll(directoriesPath, c.perm)
	if err != nil {
		return fmt.Errorf("error creating necessary directories: %w", err)
	}

	f, err := os.Create(pkPath)
	if err != nil {
		return fmt.Errorf("error creating specified pk: %w", err)
	}
	defer f.Close()

	_, err = f.Write(blob)
	if err != nil {
		return fmt.Errorf("error writing data to pk: %w", err)
	}

	return err
}

func (c CrudOs) Read(pk []string) ([]byte, error) {

	if len(pk) == 0 {
		return nil, fmt.Errorf("pk cannot be empty")
	}

	pkPath := c.pkToPath(pk)
	f, err := os.Open(pkPath)
	if err != nil {
		return nil, fmt.Errorf("error opening pk: %w", err)
	}
	defer f.Close()

	blob, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("error reading pk: %w", err)
	}

	return blob, nil
}

func (c CrudOs) Update(pk []string, blob []byte) error {

	if len(pk) == 0 {
		return fmt.Errorf("pk cannot be empty")
	}

	pkPath := c.pkToPath(pk)
	exists, err := c.Exists(pk)
	if err != nil {
		return fmt.Errorf("error verifying pk existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("cannot update contents of non-existant pk")
	}

	f, err := os.Create(pkPath)
	if err != nil {
		return fmt.Errorf("error truncating specified pk: %w", err)
	}
	defer f.Close()

	_, err = f.Write(blob)
	if err != nil {
		return fmt.Errorf("error writing data to pk: %w", err)
	}

	return nil
}

func (c CrudOs) Delete(key []string) error {

	if len(key) == 0 {
		return fmt.Errorf("pk cannot be empty")
	}

	pkPath := c.pkToPath(key)
	exists, err := c.Exists(key)
	if err != nil {
		return fmt.Errorf("error verifying pk existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("cannot delete non-existant pk")
	}

	err = os.Remove(pkPath)
	if err != nil {
		return fmt.Errorf("error deleting pk: %w", err)
	}

	dirPath := filepath.Dir(pkPath)
	isDirEmpty, err := c.isDirEmpty(dirPath)
	if err != nil {
		return fmt.Errorf("error checking if parent directory is empty: %w", err)
	}
	if isDirEmpty && dirPath != c.root {
		err := c.Delete(c.absPathToPk(dirPath))
		if err != nil {
			return fmt.Errorf("error deleting parent directory: %w", err)
		}
	}

	return nil
}

func (c CrudOs) Exists(pk []string) (bool, error) {

	if len(pk) == 0 {
		return false, fmt.Errorf("pk cannot be empty")
	}

	pkPath := c.pkToPath(pk)
	_, err := os.Stat(pkPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("unexpected error verifying pk existence: %w", err)
	}

	return true, nil
}

func (c CrudOs) SizeOf(key []string) (metrics.Byte, error) {
	if len(key) == 0 {
		return 0, fmt.Errorf("pk cannot be empty")
	}

	pkPath := c.pkToPath(key)
	fileInfo, err := os.Stat(pkPath)
	if err != nil {
		return 0, fmt.Errorf("error retrieving pk info: %w", err)
	}

	if fileInfo.IsDir() {
		dirInfo, err := metrics.MeasureDir(pkPath)
		if err != nil {
			return 0, fmt.Errorf("error measuring directory size: %w", err)
		}
		return dirInfo.Size(), nil
	}

	return metrics.Byte(fileInfo.Size()), nil
}
