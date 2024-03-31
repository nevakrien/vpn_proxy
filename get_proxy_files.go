package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var (
		numDirectories     int
		maxFilesPerCountry int
		clearProxies       bool
	)
	flag.IntVar(&numDirectories, "numDirs", 8, "Number of proxy directories to create")
	flag.IntVar(&maxFilesPerCountry, "maxFiles", 3, "Maximum number of files to take from each country")
	flag.BoolVar(&clearProxies, "clear", true, "Clear proxy directories before starting")
	flag.Parse()

	srcDir := "/etc/openvpn/ovpn_tcp/"
	authFilePath := "vpn-auth.txt"
	destBaseDir := "dockers" // Update this path

	cwd, err := os.Getwd()
    if err != nil {
        fmt.Println("Failed to get current working directory:", err)
        return
    }

    authFilePath = filepath.Join(cwd, authFilePath)

	// Optionally clear existing config files
	if clearProxies {
		clearProxiesDirs(destBaseDir)
	}

	// Create directories and nested ovpn_configs directories
	for i := 0; i < numDirectories; i++ {
		dirPath := filepath.Join(destBaseDir, fmt.Sprintf("proxy%d", i),"ovpn_configs")
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			fmt.Printf("Failed to create directory %s: %v\n", dirPath, err)
			return
		}
		dirPath=filepath.Join(destBaseDir, fmt.Sprintf("proxy%d", i),"auth.txt")
		if err := copyFile(authFilePath,dirPath); err != nil {
			fmt.Printf("Failed to create auth for %s: %v\n", dirPath, err)
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
// clearProxiesDirs clears all proxyN directories in the base directory
func clearProxiesDirs(baseDir string) {
	// List all items in the base directory
	items, err := os.ReadDir(baseDir)
	if err != nil {
		fmt.Printf("Failed to list directory %s: %v\n", baseDir, err)
		return
	}

	// Iterate through all items and remove those matching the proxyN pattern
	for _, item := range items {
		if item.IsDir() && strings.HasPrefix(item.Name(), "proxy") {
			dirPath := filepath.Join(baseDir, item.Name())
			if err := os.RemoveAll(dirPath); err != nil {
				fmt.Printf("Failed to remove directory %s: %v\n", dirPath, err)
			} else {
				fmt.Printf("Cleared configs in %s\n", dirPath)
			}
		}
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
