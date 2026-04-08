package utils

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func WriteFile(filePath string, data string) {
	err := os.WriteFile(filePath, []byte(data), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func WriteFileBytes(filename string, bytes []byte) {
	err := os.WriteFile(filename, bytes, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func ReadFile(filename string) []byte {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	return data
}

func RemoveFileByPath(filePath string) {
	err := os.Remove(filePath)
	if err != nil {
		log.Fatal(err)
	}
}

func FindFile(dir string, filename string) string {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.Name() == filename {
			return filepath.Join(dir, file.Name())
		}
	}
	return ""
}

func ReadDir(dirname string) []fs.DirEntry {
	files, err := os.ReadDir(dirname)
	if err != nil {
		log.Fatal(err)
	}
	return files
}

func FileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func DirHasFiles(dir string) bool {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	return len(files) > 0
}
