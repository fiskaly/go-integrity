package main

import (
	"flag"
	"fmt"
	"os"
	"slices"

	"github.com/fiskaly/go-integrity/pkg"
)

const Command = "go-integrity"

var Version = "v1.2.3"

func main() {
	var showVersion bool
	flag.BoolVar(&showVersion, "v", false, "print application version")

	var showHelp bool
	flag.BoolVar(&showHelp, "h", false, "print usage information")

	var includeSource bool
	flag.BoolVar(&includeSource, "s", false, "include source of <file.go> into <file.go-integrity>")

	flag.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			"Usage: %s [-h | -v | -s] <file1.go> [file2.go ...]\n",
			Command,
		)
		flag.PrintDefaults()
	}

	flag.Parse()

	if showVersion {
		fmt.Printf(
			"%s version %s\n",
			Command,
			Version,
		)
		os.Exit(0)
	}

	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	files := flag.Args()
	if len(files) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	filePaths := make([]string, 0, len(files))
	fileHashes := make(map[string]string)

	for _, filePath := range files {
		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filePath, err)
			os.Exit(1)
		}

		// Calculate the stripped implementation hash for the current file
		fileContent, fileHash, err := pkg.Hashing(content)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error hashing file %s: %v\n", filePath, err)
			os.Exit(1)
		}

		outputFile := filePath + "-integrity"
		outputData := fileHash

		if includeSource {
			outputData = outputData + "\n"
			outputData = outputData + "---\n"
			outputData = outputData + string(fileContent) + "\n"
		}

		err = os.WriteFile(outputFile, []byte(outputData), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing hash to file %s: %v\n", outputFile, err)
			os.Exit(1)
		}

		fileHashes[filePath] = fileHash
		filePaths = append(
			filePaths,
			filePath,
		)
	}

	slices.Sort(filePaths)

	for _, filePath := range filePaths {
		fmt.Printf("%s %s\n", fileHashes[filePath], filePath)
	}
}
