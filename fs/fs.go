package fs

import (
	"fmt"
	"os/exec"

	"os"
)

// osFileSystem uses afero to perform fs operations
type osFileSystem struct {
}

// FileSystem abstracts fs operations
type FileSystem interface {
	// DirExists checks for existence of directory
	DirExists(dir string) (bool, error)
	// CreateDir creates a new directory
	CreateDir(dir string, recursive bool, perm os.FileMode) error
	// RemoveDir deletes a directory
	RemoveDir(dir string, recursive bool) error
	// Mount mounts a block device
	Mount(device string, target string) error
	// Unmount unmounts a block device
	Unmount(target string) error
	// Format formats a block device
	Format(device string) error
}

// NewFileSystem creates a new fs object
func NewFileSystem() FileSystem {
	return &osFileSystem{}
}

// DirExists checks for existence of directory
func (fs *osFileSystem) DirExists(dir string) (bool, error) {
	stat, err := os.Stat(dir)

	if err == nil && stat.IsDir() {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// CreateDir creates a new directory
func (fs *osFileSystem) CreateDir(dir string, recursive bool, perm os.FileMode) error {
	if recursive {
		return os.MkdirAll(dir, perm)
	}
	return os.Mkdir(dir, perm)
}

// RemoveDir deletes a directory
func (fs *osFileSystem) RemoveDir(dir string, recursive bool) error {
	if recursive {
		return os.RemoveAll(dir)
	}
	return os.Remove(dir)
}

// Mount mounts a block device
func (fs *osFileSystem) Mount(device string, target string) error {
	return osExec("mount", "-o", "defaults,discard", device, target)
}

// Unmount unmounts a block device
func (fs *osFileSystem) Unmount(target string) error {
	return osExec("unmount", target)
}

// Format formats a block device
func (fs *osFileSystem) Format(target string) error {
	return osExec("mkfs.ext4", target)
}

// osExec runs a shell command
func osExec(cmd string, args ...string) error {
	command := exec.Command(cmd, args...)

	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("%s failed, arguments: %q\noutput: %s", cmd, args, string(output))
	}
	return nil
}
