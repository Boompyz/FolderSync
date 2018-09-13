package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
)

type fileQueueEntry struct {
	sourceFile string
	destFile   string
	size       int64
}

var bytesToCopy int64
var bytesCopied int64
var filesToCopy []fileQueueEntry

func errCheck(err error) {
	if err != nil {
		panic(err)
	}
}

// Returns true if the files are the same size
func sameFile(sourceFile, destFile os.FileInfo) bool {
	return sourceFile.Size() == destFile.Size()
}

// Copies a file
func copyFile(fileEntry fileQueueEntry) {
	fmt.Printf("Current progress: %vMB/%vMB  [%v%%].\t", bytesCopied/1000000, bytesToCopy/1000000, (bytesCopied * 100 / bytesToCopy))
	fmt.Printf("Copying %v.\n", fileEntry)

	data, err := ioutil.ReadFile(fileEntry.sourceFile)
	errCheck(err)

	err = ioutil.WriteFile(fileEntry.destFile, data, os.ModePerm)
	errCheck(err)

	bytesCopied += fileEntry.size
}

//Adds a file to the copy queue
func copyFileQueue(sourceFile, destFile string) {
	fileInfo, err := os.Stat(sourceFile)
	errCheck(err)

	newFileQueueEntry := fileQueueEntry{sourceFile, destFile, fileInfo.Size()}
	bytesToCopy += newFileQueueEntry.size

	filesToCopy = append(filesToCopy, newFileQueueEntry)
}

// Copies folder recursively
func syncFolders(sourceDir string, outDir string) {
	filesSource, err := ioutil.ReadDir(sourceDir)
	errCheck(err)

	filesDest, err := ioutil.ReadDir(outDir)
	errCheck(err)

	sort.Slice(filesSource, func(i, j int) bool { return filesSource[i].Name() < filesSource[j].Name() })
	sort.Slice(filesDest, func(i, j int) bool { return filesDest[i].Name() < filesDest[j].Name() })

	idxSource, idxDest := 0, 0

	for idxSource < len(filesSource) || idxDest < len(filesDest) {
		var nameSource, nameDest string
		if idxSource < len(filesSource) {
			nameSource = filesSource[idxSource].Name()
		} else {
			nameSource = ""
		}

		if idxDest < len(filesDest) {
			nameDest = filesDest[idxDest].Name()
		} else {
			nameDest = ""
		}

		newSource := sourceDir + "/" + nameSource
		newDest := outDir + "/" + nameSource
		oldDest := outDir + "/" + nameDest

		// If they differ, check which one comes first
		if nameSource != nameDest {
			// If nameSource is smaller than nameDest, it is possible nameDest to exist in source files
			// Copy the source to dest
			if (nameSource < nameDest || nameDest == "") && nameSource != "" {

				// If it is a dir, create and sync
				if filesSource[idxSource].IsDir() {
					os.Mkdir(newDest, os.ModePerm)
					syncFolders(newSource, newDest)
				} else {
					copyFileQueue(newSource, newDest)
				}

				idxSource++
			} else { // nameDest doesn't exists -> it shouldn't be there
				if filesDest[idxDest].IsDir() {
					os.RemoveAll(oldDest)
				} else {
					os.Remove(oldDest)
				}

				idxDest++
			}
		} else { // Names are the same
			fileSoure := filesSource[idxSource]
			fileDest := filesDest[idxDest]

			if fileSoure.IsDir() && !fileDest.IsDir() {
				// If Source is folder and dist is file
				// delete the file and sync the folder
				os.Remove(newDest)
				os.Mkdir(newDest, os.ModePerm)
				syncFolders(newSource, newDest)
			} else if !fileSoure.IsDir() && fileDest.IsDir() {
				// If Source if file an dist is folder
				// delete Dest and copy the file
				os.RemoveAll(newDest)
				copyFileQueue(newSource, newDest)
			} else if fileSoure.IsDir() && fileDest.IsDir() {
				// If they are both folders, check inside
				syncFolders(newSource, newDest)
			} else if !fileSoure.IsDir() && !fileDest.IsDir() {
				// check if they are the same
				// if not, copy
				if !sameFile(fileSoure, fileDest) {
					os.Remove(newDest)
					copyFileQueue(newSource, newDest)
				}
			}

			idxDest++
			idxSource++
		}
	}
}

func main() {

	if len(os.Args) < 3 {
		fmt.Println("Not enough arguments!")
		fmt.Println("Usage:")
		fmt.Println("foldersync <source> <dest>")
		return
	}

	bytesToCopy = 0
	filesToCopy = make([]fileQueueEntry, 0)

	sourceDir := os.Args[1]
	outDir := os.Args[2]

	syncFolders(sourceDir, outDir)

	fmt.Printf("Have to copy %v bytes (%v MB).\n", bytesToCopy, bytesToCopy/1000000)
	for _, fileEntry := range filesToCopy {
		copyFile(fileEntry)
	}
}
