package main

import (
	"fmt"
	"os"
)

// FileExists returns whether a given name exists and is a regular file.
func FileExists(name string) bool {
	if fi, err := os.Stat(name); err == nil && fi.Mode().IsRegular() {
		return true
	}
	return false
}

// FileExists returns whether a given path exists and is a directory.
func DirectoryExists(name string) bool {
	if fi, err := os.Stat(name); err == nil && fi.Mode().IsDir() {
		return true
	}
	return false
}

// ReadFile reads all the content of a file and returns it as a slice of bytes.
func ReadFile(name string) ([]byte, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	var content []byte
	buff := make([]byte, 1024)
	offset := int64(0)
	for {
		n, err := f.ReadAt(buff, offset)
		offset += int64(n)
		content = append(content, buff[:n]...)
		if err != nil {
			break
		}
	}
	return content, nil
}

// WalkDir walks a directory tree beginning at the given root.
// In every directory, it first calls the callback on every regular file.
// Then it pushes all subdirectories to the queue.
func WalkDir(root string, callback func(fname string) error) error {
	for dirQueue := []string{root}; len(dirQueue) > 0; dirQueue = dirQueue[1:] {
		dir, err := os.Open(dirQueue[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}

		names, err := dir.Readdirnames(0)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}

		for _, name := range names {
			if FileExists(dirQueue[0] + "/" + name) {
				err = callback(dirQueue[0] + "/" + name)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return err
				}
			} else if DirectoryExists(dirQueue[0] + "/" + name) {
				dirQueue = append(dirQueue, dirQueue[0]+"/"+name)
			}
		}
	}
	return nil
}
