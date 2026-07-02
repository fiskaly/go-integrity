package pkg

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
)

// ReadFirstLine loads the file from the provided filePath and returns the first line.
func ReadFirstLine(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		return scanner.Text(), nil
	}

	err = scanner.Err()
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	return "", fmt.Errorf("file is empty")
}

// stripEmptyLines removes all lines that are empty or contain only whitespace.
func stripEmptyLines(src []byte) []byte {
	scanner := bufio.NewScanner(bytes.NewReader(src))
	var buf bytes.Buffer

	for scanner.Scan() {
		line := scanner.Text()
		// only keep the line if it has actual content
		if strings.TrimSpace(line) != "" {
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
	}

	return buf.Bytes()
}
