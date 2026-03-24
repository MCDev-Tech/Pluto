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
	if before, ok := strings.CutSuffix(classPath, ".class"); ok {
		classPath = before
	}
	return classPath
}

func ExtractClassFromJar(jarPath, classPath, outPath string) ([]string, error) {
	if jarPath == "" || classPath == "" || outPath == "" {
		return nil, errors.New("jarPath, classPath, outPath cannot be empty")
	}

	normalized := normalizeClassPath(classPath)
	// 构建匹配前缀（主类完整路径 + $），用于匹配内部类
	matchPrefix := normalized + "$"
	// 主类的完整文件名（含.class）
	mainClassName := normalized + ".class"

	zipReader, err := zip.OpenReader(jarPath)
	if err != nil {
		return nil, err
	}
	defer zipReader.Close()

	// 收集符合条件的zip文件条目
	var targetEntries []*zip.File
	for _, f := range zipReader.File {
		// 匹配主类 或 内部类（同目录且以主类$开头）
		if f.Name == mainClassName || strings.HasPrefix(f.Name, matchPrefix) {
			// 额外校验内部类和主类同目录
			if f.Name != mainClassName {
				mainClassDir := filepath.Dir(mainClassName)
				entryDir := filepath.Dir(f.Name)
				if mainClassDir != entryDir {
					continue
				}
			}
			targetEntries = append(targetEntries, f)
		}
	}

	if len(targetEntries) == 0 {
		return nil, errors.New("no class found in jar: " + mainClassName + " (including inner classes)")
	}

	// 确保输出目录存在
	folder := filepath.Join(outPath, filepath.Dir(classPath))
	if err := os.MkdirAll(folder, os.ModePerm); err != nil {
		return nil, err
	}

	// 存储提取成功的文件路径
	extractedFiles := make([]string, 0, len(targetEntries))

	// 遍历提取所有符合条件的class文件
	for _, entry := range targetEntries {
		// 打开zip内的文件
		inFile, err := entry.Open()
		if err != nil {
			return extractedFiles, err
		}

		// 构建输出文件路径（输出目录 + 原文件名）
		outFilePath := filepath.Join(outPath, entry.Name)

		// 创建输出文件
		outFile, err := os.Create(outFilePath)
		if err != nil {
			inFile.Close()
			return extractedFiles, err
		}

		// 复制文件内容
		_, err = io.Copy(outFile, inFile)
		if err != nil {
			inFile.Close()
			outFile.Close()
			return extractedFiles, err
		}

		outFile.Sync()
		outFile.Close()
		inFile.Close()

		// 记录成功提取的文件路径
		absPath, err := filepath.Abs(outFilePath)
		if err != nil {
			inFile.Close()
			outFile.Close()
			return extractedFiles, err
		}
		extractedFiles = append(extractedFiles, absPath)
	}

	return extractedFiles, nil
}
