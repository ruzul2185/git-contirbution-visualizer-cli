package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// getDotFilePath return the dot file for the repos list.
// Creates it and the enclosing folder if it does not exist.
func getDotFilePath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	dotFile := filepath.Join(usr.HomeDir, ".gogitlocalstats")

	return dotFile
}

func recursiveScanFolder(folder string) []string {
	return scanGitFolders(make([]string, 0), folder)
}

// scanGitFolders returns a list of subfolders of `folder` ending with `.git`.
// Returns the base folder of the repo, the .git folder parent.
// Recursively searches in the subfolders by passing an existing `folders` slice.
func scanGitFolders(folders []string, folder string) []string {
	// trim the last `/`
	folder = strings.TrimSuffix(folder, "/")

	f, err := os.Open(folder)
	if err != nil {
		log.Fatal(err)
	}
	files, err := f.ReadDir(-1)
	f.Close()
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			path := filepath.Join(folder, file.Name())
			if file.Name() == ".git" {
				path = strings.TrimSuffix(path, "/.git")
				fmt.Println(path)
				folders = append(folders, path)
				continue
			}
			if file.Name() == "vendor" || file.Name() == "node_modules" {
				continue
			}
			folders = scanGitFolders(folders, path)
		}
	}

	return folders
}

// addNewSliceElementsToFile given a slice of strings representing paths, stores them
// to the filesystem
func addNewSliceElementsToFile(filePath string, newRepos []string) {
	existingRepos := parseFileLinesToSlice(filePath)
	repos := joinSlices(newRepos, existingRepos)
	dumpStringsSliceToFile(repos, filePath)
}

// parseFileLinesToSlice given a file path string, gets the content
// of each line and parses it to a slice of strings.
func parseFileLinesToSlice(filePath string) []string {
	f := openFile(filePath)
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			panic(err)
		}
	}

	return lines
}

// openFile opens the file located at `filePath`. Creates it if not existing.
func openFile(filePath string) *os.File {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_RDWR, 0755)
	if err != nil {
		if os.IsNotExist(err) {
			// file does not exist
			_, err := os.Create(filePath)
			if err != nil {
				panic(err)
			}
		} else {
			//other error
			panic(err)
		}
	}
	return f
}

// joinSlices adds teh element of the `new` slice
// into the `existing` slice, only if not already there
func joinSlices(new []string, existing []string) []string {
	seen := make(map[string]struct{})

	for _, r := range existing {
		seen[r] = struct{}{}
	}

	for _, r := range new {
		if _, ok := seen[r]; !ok {
			existing = append(existing, r)
			seen[r] = struct{}{}
		}
	}

	return existing
}

// dumpStringsSliceToFile writes content to the file in path `filePath` (overwrite existing content)
func dumpStringsSliceToFile(repos []string, filePath string) {
	content := strings.Join(repos, "\n")
	os.WriteFile(filePath, []byte(content), 0755)
}
