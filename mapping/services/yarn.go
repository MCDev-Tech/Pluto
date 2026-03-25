package services

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"pluto/global"
	"pluto/mapping/java"
	"pluto/util"
	"pluto/util/network"
	"pluto/vanilla"
	"strings"
)

type Yarn struct{}

type YarnVersion struct {
	GameVersion string `json:"gameVersion"`
	Separator   string `json:"separator"`
	Build       int    `json:"build"`
	Maven       string `json:"maven"`
	Version     string `json:"version"`
	Stable      bool   `json:"stable"`
}

var yarnMappings = make(map[string]*java.Mappings)

func (s *Yarn) GetName() string {
	return "yarn"
}

func (s *Yarn) GetPathOrDownload(mcVersion string) (string, error) {
	path := global.GetMappingPath(s, mcVersion, "tiny")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return path, nil
	}
	body, err := network.Get(global.Config.Urls.FabricMeta + "/v2/versions/yarn")
	if err != nil {
		return "", errors.New("Unable to download yarn versions: " + err.Error())
	}
	var versions []YarnVersion
	if err := json.Unmarshal(body, &versions); err != nil {
		return "", errors.New("Unable to unmarshal yarn versions: " + err.Error())
	}
	var latestVersion *YarnVersion
	for i := range versions {
		version := &versions[i]
		if version.GameVersion == mcVersion {
			if latestVersion == nil || version.Build > latestVersion.Build {
				latestVersion = version
			}
		}
	}
	if latestVersion == nil {
		return "", errors.New("Unable to find latest version for " + mcVersion)
	}
	jar, err := network.Get(fmt.Sprintf(global.Config.Urls.FabricMaven+"/net/fabricmc/yarn/%s/yarn-%s-mergedv2.jar", latestVersion.Version, latestVersion.Version))
	if err != nil {
		return "", errors.New("Unable to download yarn mapping: " + err.Error())
	}
	data, err := ExtractMappingsTinyFromJar(jar)
	if err != nil {
		return "", errors.New("Unable to unzip yarn mapping: " + err.Error())
	}
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (s *Yarn) GetMappingCacheOrError(mcVersion string) (*java.Mappings, error) {
	if mapping, ok := yarnMappings[mcVersion]; ok {
		return mapping, nil
	}
	return nil, errors.New("not cached yet")
}

func (s *Yarn) SaveMappingCache(mcVersion string, mapping *java.Mappings) {
	yarnMappings[mcVersion] = mapping
}

func (s *Yarn) LoadMapping(mcVersion string) (*map[java.SingleInfo]java.SingleInfo, error) {
	path, err := s.GetPathOrDownload(mcVersion)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件: %w", err)
	}
	defer file.Close()

	mapping := make(map[java.SingleInfo]java.SingleInfo)
	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		_ = scanner.Text() //Remove the first definition line
	}
	notchClassCache, yarnClassCache := "", ""
	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " \t")
		split := strings.Split(line, "\t")
		if len(split) < 4 {
			continue
		}
		switch split[0] {
		case "c":
			notchClassCache, yarnClassCache = split[1], split[3]
			notch, yarn := java.PackClassInfo(notchClassCache), java.PackClassInfo(yarnClassCache)
			mapping[notch] = yarn
		case "m":
			notch, yarn := java.PackMethodInfo(split[2], notchClassCache, split[1]), java.PackMethodInfo(split[4], yarnClassCache, "")
			mapping[notch] = yarn
		case "f":
			notch, yarn := java.PackFieldInfo(split[2], notchClassCache, split[1]), java.PackFieldInfo(split[4], yarnClassCache, "")
			mapping[notch] = yarn
		}
		//case "p":
	}
	//Post processor
	result, classMapping := make(map[java.SingleInfo]java.SingleInfo), make(map[string]string)
	for notch, yarn := range mapping {
		if notch.Type == "class" {
			classMapping[notch.Signature] = yarn.Signature
		}
	}
	for notch, yarn := range mapping {
		notch.Name = java.FullToClassName(notch.Name)
		yarn.Name = java.FullToClassName(yarn.Name)
		switch notch.Type {
		case "method":
			yarn.Signature = java.ObfuscateMethodSignature(notch.Signature, classMapping)
		case "field":
			yarn.Signature = java.ObfuscateTypeSignature(notch.Signature, classMapping)
		}
		result[notch] = yarn
	}
	return &result, nil
}

func (s *Yarn) Remap(mcVersion string) (string, error) {
	jarPath, err := vanilla.GetMcJarPath(mcVersion)
	if err != nil {
		return "", err
	}
	mappingPath, err := s.GetPathOrDownload(mcVersion)
	if err != nil {
		return "", err
	}
	outputPath := global.GetRemappedPath(s, mcVersion)
	err = util.ExecuteCommand(global.Config.JavaPath, []string{"-cp", global.ClassPath, global.TinyRemapperMainClass, jarPath, outputPath, mappingPath, "official", "named"}, false)
	if err != nil {
		return "", err
	}
	return outputPath, nil
}

func ExtractMappingsTinyFromJar(jarData []byte) ([]byte, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(jarData), int64(len(jarData)))
	if err != nil {
		return nil, fmt.Errorf("解析JAR/ZIP格式失败: %w", err)
	}
	targetPath := "mappings/mappings.tiny"
	for _, file := range zipReader.File {
		if file.Name == targetPath {
			fileReader, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("打开 mappings.tiny 文件失败: %w", err)
			}
			defer fileReader.Close()

			fileContent, err := io.ReadAll(fileReader)
			if err != nil {
				return nil, fmt.Errorf("读取 mappings.tiny 内容失败: %w", err)
			}
			return fileContent, nil
		}
	}
	return nil, fmt.Errorf("JAR中未找到文件: %s", targetPath)
}
