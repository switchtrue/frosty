package artifact

import (
	"archive/zip"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func MakeArtifactArchive(artifactDir string, target string) bool {
	artifactFiles := listArtifactFiles(artifactDir, target)

	if len(artifactFiles) == 0 {
		return false
	}
	makeZipFromFiles(target, artifactFiles, artifactDir)
	return true
}

func listArtifactFiles(artifactDir string, target string) []string {
	var artifactFiles []string

	err := filepath.Walk(artifactDir, func(path string, f os.FileInfo, err error) error {
		if path != artifactDir && !isDirectory(path) && path != target {
			artifactFiles = append(artifactFiles, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("filepath.Walk() returned %v\n", err)
	}

	return artifactFiles
}

func isDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}
	return fileInfo.IsDir()
}

func makeZipFromFiles(target string, sourceFileList []string, basePath string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	// Create a new zip archive.
	w := zip.NewWriter(zipfile)
	defer w.Close()

	for _, file := range sourceFileList {
		relativeFileName, err := filepath.Rel(basePath, file)
		if err != nil {
			log.Fatal(err)
		}

		f, err := w.Create(relativeFileName)
		if err != nil {
			log.Fatal(err)
		}

		fileContents, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}

		_, err = f.Write(fileContents)
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}
