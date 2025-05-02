package fileutils

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

func MkdirAll(fs afero.Fs, path string, perm os.FileMode) error {
	base := "/"
	for dir := path; dir != "/"; dir = filepath.Dir(dir) {
		exists, _ := afero.DirExists(fs, dir)
		if exists {
			base = dir
			break
		}
	}
	err := fs.MkdirAll(path, perm)
	if err != nil {
		return err
	}
	for dir := path; dir != base; dir = filepath.Dir(dir) {
		fs.Chmod(dir, perm)
	}
	return nil
}

// CopyDir copies a directory from source to dest and all
// of its sub-directories. It doesn't stop if it finds an error
// during the copy. Returns an error if any.
func CopyDir(fs afero.Fs, source, dest string) error {
	// Get properties of source.
	srcinfo, err := fs.Stat(source)
	if err != nil {
		return err
	}

	// Create the destination directory.
	err = MkdirAll(fs, dest, srcinfo.Mode())
	if err != nil {
		return err
	}

	dir, _ := fs.Open(source)
	obs, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	var errs []error

	for _, obj := range obs {
		fsource := source + "/" + obj.Name()
		fdest := dest + "/" + obj.Name()

		if obj.IsDir() {
			// Create sub-directories, recursively.
			err = CopyDir(fs, fsource, fdest)
			if err != nil {
				errs = append(errs, err)
			}
		} else {
			// Perform the file copy.
			err = CopyFile(fs, fsource, fdest)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	var errString string
	for _, err := range errs {
		errString += err.Error() + "\n"
	}

	if errString != "" {
		return errors.New(errString)
	}

	return nil
}
