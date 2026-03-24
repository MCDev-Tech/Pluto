package util

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func normalizeClassPath(classPath string) string {
	classPath = strings.TrimSpace(classPath)
	classPath = strings.Trim(classPath, "/")
	classPath = strings.ReplaceAll(classPath, ".", "/")
	if strings.HasSuffix(classPath, ".class") {
		classPath = strings.TrimSuffix(classPath, ".class")
	}
	return classPath
}

func ExtractClassFromJar(jarPath, classPath, outPath string) error {
	if jarPath == "" || classPath == "" || outPath == "" {
		return errors.New("jarPath, classPath, outPath cannot be empty")
	}

	normalized := normalizeClassPath(classPath)
	zipReader, err := zip.OpenReader(jarPath)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	targetName := normalized + ".class"
	var entry *zip.File
	for _, f := range zipReader.File {
		if f.Name == targetName {
			entry = f
			break
		}
	}
	if entry == nil {
		return errors.New("class not found in jar: " + targetName)
	}

	if err := os.MkdirAll(filepath.Dir(outPath), os.ModePerm); err != nil {
		return err
	}
	inFile, err := entry.Open()
	if err != nil {
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, inFile)
	return err
}
