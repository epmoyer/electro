package electro

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func pathExistsFS(fsys fs.FS, path string) bool {
	_, err := fs.Stat(fsys, path)
	return !os.IsNotExist(err)
}

func pathIsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		// Handles both os.ErrNotExist and any other errors
		return false
	}
	return info.IsDir()
}

func pathIsDirFS(fsys fs.FS, path string) bool {
	info, err := fs.Stat(fsys, path)
	if err != nil {
		// Handles both fs.ErrNotExist and any other errors
		return false
	}
	return info.IsDir()
}

func pathIsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		// Handles both os.ErrNotExist and any other errors
		return false
	}
	return !info.IsDir()
}

func pathIsFileFS(fsys fs.FS, path string) bool {
	info, err := fs.Stat(fsys, path)
	if err != nil {
		// Handles both fs.ErrNotExist and any other errors
		return false
	}
	return !info.IsDir()
}

// readFileLines reads data from a text file and returns a slice of file lines
func readFileLines(pathFile string) ([]string, error) {
	file, err := os.Open(pathFile)
	if err != nil {
		return []string{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// We set a large token sixe because we will be parsing HTML with
	// large (typically ~60K) inlined base64 data which appears on a
	// single line.
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return []string{}, err
	}
	return lines, nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Ensure destination directory exists
	err = os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}

func copyFileFromFS(fsysSrc fs.FS, src, dst string) error {
	sourceFile, err := fsysSrc.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Ensure destination directory exists
	err = os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}

func copyDirectoryContents(srcDir, dstDir string) error {
	// Check if source directory exists
	if !pathIsDir(srcDir) {
		return fmt.Errorf("source directory does not exist: %s", srcDir)
	}

	// Create destination directory
	err := os.MkdirAll(dstDir, 0755)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		if entry.IsDir() {
			err = copyDirectoryContents(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func copyDirectoryContentsFromFS(fsysSrc fs.FS, srcDir, dstDir string) error {
	// Check if source directory exists
	if !pathIsDirFS(fsysSrc, srcDir) {
		return fmt.Errorf("source directory does not exist: %s", srcDir)
	}

	// Create destination directory
	err := os.MkdirAll(dstDir, 0755)
	if err != nil {
		return err
	}

	entries, err := fs.ReadDir(fsysSrc, srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		if entry.IsDir() {
			err = copyDirectoryContentsFromFS(fsysSrc, srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = copyFileFromFS(fsysSrc, srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func writeStringToFileEnsureDir(pathFile string, content string) error {
	// Ensure directory exists
	err := os.MkdirAll(filepath.Dir(pathFile), 0755)
	if err != nil {
		return err
	}

	// Write content to file
	return os.WriteFile(pathFile, []byte(content), 0644)
}
