package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
)

// IsDir checks to see if passed in path is a directory
func IsDir(path string) (isdir bool) {
	dir, err := os.Stat(path)
	if err != nil {
		return false
	}

	return dir.IsDir()
}

// IsFile checks to see if the passed in path is a file
func IsFile(path string) (isFile bool) {
	file, err := os.Stat(path)
	if err != nil {
		return false
	}

	return file.Mode().IsRegular()
}

// CopyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file. The file mode will be copied from the source and
// the copied data is synced/flushed to stable storage.
func CopyFile(src string, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	logger.Debugf("\tName: %s", out.Name())

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	logger.Debugf("\t\tMode: %s", si.Mode())

	err = os.Chown(dst,
		int(si.Sys().(*syscall.Stat_t).Uid),
		int(si.Sys().(*syscall.Stat_t).Gid))
	if err != nil {
		return
	}

	logger.Debugf("\t\tOwner: %d.%d",
		int(si.Sys().(*syscall.Stat_t).Uid),
		int(si.Sys().(*syscall.Stat_t).Gid))

	return
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
// Symlinks are ignored and skipped.
func CopyDir(src string, dst string) (err error) {
	syscall.Umask(0)

	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		logger.Warningf("Source directory given %s is not a directory.\n", src)
		return fmt.Errorf("source is not a directory")
	}

	di, err := os.Lstat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}

	if di != nil && di.Mode()&os.ModeSymlink != 0 {
		logger.Infof("Destination directory %s is a symlink.\n", dst)
		return
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	err = os.Chown(dst,
		int(si.Sys().(*syscall.Stat_t).Uid),
		int(si.Sys().(*syscall.Stat_t).Gid))
	if err != nil {
		return
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}

// LinkShared attempts to resolve the shared locations to actual files and
// links them to the source locations.
func LinkShared(shared []string, source string, destination string) (err error) {

	for _, entry := range shared {
		sourcePath := filepath.Join(source, entry)
		destinationPath := filepath.Join(destination, entry)

		logger.Infof("Linking %s as %s\n", entry, destinationPath)

		absSourcePath, err := filepath.Abs(sourcePath)
		if err != nil {
			return err
		}

		err = os.RemoveAll(destinationPath)
		if err != nil {
			return err
		}

		if IsFile(absSourcePath) || IsDir(absSourcePath) {
			os.Symlink(absSourcePath, destinationPath)
		}
	}

	return
}
