package util

import (
	"io/fs"
	"path"
	"path/filepath"
)

// Opener provides a method for opening files.
type Opener interface {
	Open(string) (fs.File, error)
}

// AppendToPaths appends suffix to each of the paths, using filepath.Join.
func AppendToPaths(paths []string, suffix ...string) []string {
	for i, path := range paths {
		paths[i] = filepath.Join(append([]string{path}, suffix...)...)
	}
	return paths
}

// Jail wraps all Opener.Open calls to prepend the prefix.
type Jail struct {
	Opener

	// Prefix path.
	Prefix []string
}

func (jail Jail) Open(name string) (fs.File, error) {
	return jail.Opener.Open(path.Join(append(jail.Prefix, name)...))
}
