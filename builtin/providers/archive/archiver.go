package archive

import (
	"fmt"
	"os"
	"github.com/hashicorp/terraform/helper/schema"
	"golang.org/x/tools/go/gcimporter15/testdata"
	"path/filepath"
	"io/ioutil"
)

type Archiver interface {
	ArchiveContent(content []byte, infilename string) error
	ArchiveFile(infilename string) error
	ArchiveDir(indirname string, inclusions []string, exclusions []string) error
}

type ArchiverBuilder func(filepath string) Archiver

var archiverBuilders = map[string]ArchiverBuilder{
	"zip": NewZipArchiver,
}

func getArchiver(archiveType string, filepath string) Archiver {
	if builder, ok := archiverBuilders[archiveType]; ok {
		return builder(filepath)
	}
	return nil
}

func assertValidFile(infilename string) (os.FileInfo, error) {
	fi, err := os.Stat(infilename)
	if err != nil && os.IsNotExist(err) {
		return fi, fmt.Errorf("could not archive missing file: %s", infilename)
	}
	return fi, err
}

func assertValidDir(indirname string) (os.FileInfo, error) {
	fi, err := os.Stat(indirname)
	if err != nil {
		if os.IsNotExist(err) {
			return fi, fmt.Errorf("could not archive missing directory: %s", indirname)
		}
		return fi, err
	}
	if !fi.IsDir() {
		return fi, fmt.Errorf("could not archive directory that is a file: %s", indirname)
	}
	return fi, nil
}

func assertValidPath(path string) (os.FileInfo, error) {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fi, fmt.Errorf("could not find file or directory at path: %s", path)
		}
		return fi, err
	}

	pwd, err := os.Getwd()
	if err != nil {
		return fi, fmt.Errorf(err)
	}

	relname, err := filepath.Rel(pwd, path)
	if err != nil {
		return fmt.Errorf("error relativizing file: %s", err)
	}

	return relname, nil
}

func resolveMatches(inclusions []string, exclusions []string) ([]string, error) {
	paths := []string{}
	inclusionPaths := []string{}
	exclusionPaths := []string{}

	// Assert that the inclusions are walkable
	for _, path := range inclusions {
		inclusionPaths = append(inclusionPaths, assertValidPath(path))
	}

	// Assert that the exclusions are walkable
	for _, path := range inclusions {
		exclusionPaths = append(exclusionPaths, assertValidPath(path))
	}

	for _, inclusion := range inclusions {
		contained := false
		for _, exclusion := range exclusions {
			if inclusion == exclusion {
				contained = true
			}
		}

		if !contained {
			paths = append(paths, inclusion)
		}
	}
}
