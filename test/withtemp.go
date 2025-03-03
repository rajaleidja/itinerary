package test

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"
)

func withTempFiles(dir string, amount int, f func(files ...*os.File)) (err error) {
	files := make([]*os.File, 0, amount)

	for ; amount > 0; amount-- {
		file, e := os.CreateTemp(dir, "temp_"+getUniqueName())
		if e != nil {
			return e
		}

		files = append(files, file)
	}

	f(files...)

	// Cleanup after execution
	defer func() {
		// Handle any panics
		if e := recover(); e != nil {
			switch x := e.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New(fmt.Sprint(x))
			}
		}

		// Cleanup temporary files
		for _, file := range files {
			removeFile(file)
		}
	}()

	return
}

func withTempFile(dir string, f func(file *os.File)) error {
	return withTempFiles(dir, 1, func(files ...*os.File) {
		f(files[0])
	})
}

func withTempFile2(dir string, f func(file1, file2 *os.File)) error {
	return withTempFiles(dir, 2, func(files ...*os.File) {
		f(files[0], files[1])
	})
}

func withTempFile3(dir string, f func(file1, file2, file3 *os.File)) error {
	return withTempFiles(dir, 3, func(files ...*os.File) {
		f(files[0], files[1], files[2])
	})
}

func withMockFiles(dir, mockInput, mockLookup string, f func(inputFile, outputFile, lookupFile *os.File)) error {
	return withTempFile3(dir, func(f1, f2, f3 *os.File) {
		if _, err := f1.WriteString(mockInput); err != nil {
			panic(err)
		}
		_ = f1.Close()
		if _, err := f3.WriteString(mockLookup); err != nil {
			panic(err)
		}
		_ = f2.Close()

		f(f1, f2, f3)
	})
}

func getUniqueName() string {
	return strconv.FormatInt(time.Now().UTC().UnixNano(), 16)
}

func removeFileByPath(path string) bool {
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		fmt.Printf("Failed to remove temporary file '%s': %s\n", path, err)
		return false
	}
	return true
}

func removeFile(file *os.File) bool {
	// Close handle
	// If error then it's closed already
	file.Close()
	return removeFileByPath(file.Name())
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func writeAndCloseFile(t *testing.T, file *os.File, content string) {
	if _, err := file.WriteString(content); err != nil {
		t.Fatalf("Error writing to file\n%s", err)
	}
	if err := file.Close(); err != nil {
		if !errors.Is(err, os.ErrClosed) {
			t.Fatalf("Error closing file\n%s", err)
		}
	}
}
