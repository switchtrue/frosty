package artifact

import (
	"archive/zip"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func MakeArtifactArchive(artifactDir string, target string) (bool, error) {
	artifactFiles, err := listArtifactFiles(artifactDir, target)
	if err != nil {
		return false, err
	}

	if len(artifactFiles) == 0 {
		return false, nil
	}

	err = makeZipFromFiles(target, artifactFiles, artifactDir)
	if err != nil {
		return false, err
	}

	return true, nil
}

func listArtifactFiles(artifactDir string, target string) ([]string, error) {
	var artifactFiles []string

	err := filepath.Walk(artifactDir, func(path string, f os.FileInfo, err error) error {
		id, err := isDirectory(path)
		if err != nil {
			return err
		}
		if path != artifactDir && !id && path != target {
			artifactFiles = append(artifactFiles, path)
		}
		return nil
	})

	if err != nil {
		return artifactFiles, err
	}

	return artifactFiles, nil
}

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			em := fmt.Sprintf("Error determining if \"%s\" is a directory:\n%s\n", path, err)
			return false, errors.New(em)
		}
	}
	return fileInfo.IsDir(), nil
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
			return err
		}

		f, err := w.Create(relativeFileName)
		if err != nil {
			return err
		}

		fileContents, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}

		_, err = f.Write(fileContents)
		if err != nil {
			return err
		}
	}

	return nil
}
