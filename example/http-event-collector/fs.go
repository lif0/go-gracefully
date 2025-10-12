package main

import (
	"io"
	"os"
)

type fileWriter struct {
	filePath string
}

func (fw *fileWriter) WriteToDisk(data []string) error {
	file, err := os.OpenFile(fw.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, row := range data {
		_, err := io.WriteString(file, row+"\n")
		if err != nil {
			return err
		}
	}
	return nil
}
