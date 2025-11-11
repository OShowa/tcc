package pk

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// a PK (primary key) mimics directory hierarchy. It can represent a directory or a file.
type PK []string

func New(path string) PK {
	return strings.Split(path, string(os.PathSeparator))
}

func (pk PK) String() string {
	return strings.Join(pk, string(filepath.Separator))
}

func (pk PK) Prefix(prefix PK) PK {
	return slices.Concat(prefix, pk)
}

func (pk PK) Suffix(suffix PK) PK {
	return slices.Concat(pk, suffix)
}

func (pk PK) Path() string {
	return strings.Join(pk, string(os.PathSeparator))
}
