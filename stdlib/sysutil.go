package stdlib

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// File I/O utilities for Kylix

// FileOpenMode represents file open modes
type FileOpenMode int

const (
	FmRead      FileOpenMode = 0
	FmWrite     FileOpenMode = 1
	FmAppend    FileOpenMode = 2
	FmReadWrite FileOpenMode = 3
)

// TTextFile wraps a file handle for Pascal-style file I/O
type TTextFile struct {
	handle *os.File
	path   string
	mode   FileOpenMode
}

// FileOpen opens a file with the given mode
func FileOpen(path string, mode FileOpenMode) (*TTextFile, error) {
	var flag int
	switch mode {
	case FmRead:
		flag = os.O_RDONLY
	case FmWrite:
		flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	case FmAppend:
		flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	case FmReadWrite:
		flag = os.O_RDWR | os.O_CREATE
	}

	f, err := os.OpenFile(path, flag, 0644)
	if err != nil {
		return nil, fmt.Errorf("FileOpen: %w", err)
	}

	return &TTextFile{handle: f, path: path, mode: mode}, nil
}

// FileClose closes the file
func (tf *TTextFile) FileClose() error {
	if tf.handle != nil {
		return tf.handle.Close()
	}
	return nil
}

// FileReadLine reads a single line from the file
func (tf *TTextFile) FileReadLine() (string, error) {
	buf := make([]byte, 0)
	tmp := make([]byte, 1)
	for {
		n, err := tf.handle.Read(tmp)
		if n > 0 {
			if tmp[0] == '\n' {
				break
			}
			if tmp[0] != '\r' {
				buf = append(buf, tmp[0])
			}
		}
		if err != nil {
			if err == io.EOF && len(buf) > 0 {
				break
			}
			return string(buf), err
		}
	}
	return string(buf), nil
}

// FileReadAll reads the entire file content
func (tf *TTextFile) FileReadAll() (string, error) {
	data, err := io.ReadAll(tf.handle)
	if err != nil {
		return "", fmt.Errorf("FileReadAll: %w", err)
	}
	return string(data), nil
}

// FileWriteLine writes a line to the file
func (tf *TTextFile) FileWriteLine(line string) error {
	_, err := tf.handle.WriteString(line + "\n")
	return err
}

// FileWrite writes string to the file without newline
func (tf *TTextFile) FileWrite(s string) error {
	_, err := tf.handle.WriteString(s)
	return err
}

// FileEOF checks if end of file is reached
func (tf *TTextFile) FileEOF() bool {
	tmp := make([]byte, 1)
	_, err := tf.handle.Read(tmp)
	if err == io.EOF {
		return true
	}
	if err == nil {
		// Seek back one byte
		tf.handle.Seek(-1, io.SeekCurrent)
	}
	return false
}

// Convenience functions

// ReadFile reads entire file content as string
func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("ReadFile: %w", err)
	}
	return string(data), nil
}

// WriteFile writes string content to a file
func WriteFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// AppendFile appends string content to a file
func AppendFile(path string, content string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("AppendFile: %w", err)
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// CreateDir creates a directory (and parents)
func CreateDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// DeleteFile deletes a file
func DeleteFile(path string) error {
	return os.Remove(path)
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("CopyFile: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("CopyFile: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// ListDir lists files in a directory
func ListDir(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("ListDir: %w", err)
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names, nil
}

// ListFiles lists files matching a glob pattern
func ListFiles(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

// GetFileSize returns file size in bytes
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GetWorkingDir returns the current working directory
func GetWorkingDir() (string, error) {
	return os.Getwd()
}

// SetWorkingDir changes the current working directory
func SetWorkingDir(path string) error {
	return os.Chdir(path)
}

// GetTempDir returns the system temp directory
func GetTempDir() string {
	return os.TempDir()
}

// GetEnv gets an environment variable
func GetEnv(key string) string {
	return os.Getenv(key)
}

// SetEnv sets an environment variable
func SetEnv(key, value string) error {
	return os.Setenv(key, value)
}

// PathJoin joins path components
func PathJoin(parts ...string) string {
	return filepath.Join(parts...)
}

// PathDir returns the directory part of a path
func PathDir(path string) string {
	return filepath.Dir(path)
}

// PathBase returns the base name of a path
func PathBase(path string) string {
	return filepath.Base(path)
}

// PathExt returns the extension of a path
func PathExt(path string) string {
	return filepath.Ext(path)
}

// ReadLines reads a file and returns all lines as a slice
func ReadLines(path string) ([]string, error) {
	content, err := ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	return lines, nil
}

// WriteLines writes a slice of strings as lines to a file
func WriteLines(path string, lines []string) error {
	content := strings.Join(lines, "\n") + "\n"
	return WriteFile(path, content)
}
