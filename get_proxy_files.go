package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	var (
		numDirectories     int
		maxFilesPerCountry int
		clearConfigs       bool
	)
	flag.IntVar(&numDirectories, "numDirs", 8, "Number of proxy directories to create")
	flag.IntVar(&maxFilesPerCountry, "maxFiles", 3, "Maximum number of files to take from each country")
	flag.BoolVar(&clearConfigs, "clear", true, "Clear config files in each directory before starting")
	flag.Parse()

	srcDir := "/etc/openvpn/ovpn_tcp/"
	destBaseDir := "dockers" // Update this path

	// Optionally clear existing config files
	if clearConfigs {
		clearConfigFiles(destBaseDir, numDirectories)
	}

	// Create directories and nested ovpn_configs directories
	for i := 0; i < numDirectories; i++ {
		dirPath := filepath.Join(destBaseDir, fmt.Sprintf("proxy%d/ovpn_configs", i))
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			fmt.Printf("Failed to create directory %s: %v\n", dirPath, err)
			return
		}
	}

	files, err := os.ReadDir(srcDir)
	if err != nil {
		fmt.Println("Error reading source directory:", err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		countryCode := file.Name()[:2]
		destDirIndex := hashCountryCode(countryCode) % numDirectories
		destDir := filepath.Join(destBaseDir, fmt.Sprintf("proxy%d/ovpn_configs", destDirIndex))

		srcPath := filepath.Join(srcDir, file.Name())
		destPath := filepath.Join(destDir, file.Name())

		if err := copyFile(srcPath, destPath); err != nil {
			fmt.Printf("Failed to copy %s to %s: %v\n", file.Name(), destPath, err)
		} else {
			fmt.Printf("Copied %s to %s\n", file.Name(), destPath)
		}
	}
}

// clearConfigFiles clears existing config files in the proxy directories
func clearConfigFiles(baseDir string, numDirs int) {
	for i := 0; i < numDirs; i++ {
		dirPath := filepath.Join(baseDir, fmt.Sprintf("proxy%d/ovpn_configs", i))
		os.RemoveAll(dirPath) // Be careful with this
		fmt.Printf("Cleared configs in %s\n", dirPath)
	}
}

// hashCountryCode provides a simple hashing mechanism for country codes to distribute files
func hashCountryCode(countryCode string) int {
	hash := 0
	for _, char := range countryCode {
		hash += int(char)
	}
	return hash
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
