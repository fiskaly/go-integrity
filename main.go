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

const (
	SuccessEmoji = "✅"
	ErrorEmoji   = "❌"
	UpdateEmoji  = "🔄"
)

func main() {
	var showVersion bool
	flag.BoolVar(&showVersion, "v", false, "print application version")

	var showHelp bool
	flag.BoolVar(&showHelp, "h", false, "print usage information")

	var includeSource bool
	flag.BoolVar(&includeSource, "s", false, "append normalized source of <file.go> to <file.go-integrity>")

	var checkIntegrity bool
	flag.BoolVar(&checkIntegrity, "c", false, "perform validation checks if <file.go> still matches <file.go-integrity>")

	flag.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			"Usage: %s [-h | -v | -s | -c] <file1.go> [file2.go ...]\n",
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
	results := make(map[string]string)
	checkIntegrityErrors := 0

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

		fileHashes[filePath] = fileHash
		filePaths = append(
			filePaths,
			filePath,
		)

		if checkIntegrity {
			currentHash, err := pkg.ReadFirstLine(outputFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}

			if currentHash == fileHash {
				results[filePath] = SuccessEmoji
			} else {
				results[filePath] = ErrorEmoji
				checkIntegrityErrors++
			}

			continue
		}

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

		results[filePath] = UpdateEmoji
	}

	slices.Sort(filePaths)

	for _, filePath := range filePaths {
		fmt.Printf("%s %s %s\n", results[filePath], fileHashes[filePath], filePath)
	}

	if checkIntegrityErrors != 0 {
		os.Exit(-1)
	}
}
