package electro

import (
	"bufio"
	"io/fs"
	"os"
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
