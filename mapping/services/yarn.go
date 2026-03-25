package services

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
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
	jar, err := network.Get(fmt.Sprintf(global.Config.Urls.FabricMaven+"/net/fabricmc/yarn/%s/yarn-%s-tiny.gz", latestVersion.Version, latestVersion.Version))
	if err != nil {
		return "", errors.New("Unable to download yarn mapping: " + err.Error())
	}
	data, err := getMappingsTinyFromGzip(jar)
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
	for scanner.Scan() {
		line := scanner.Text()
		split := strings.Split(line, "\t")
		if len(split) < 4 {
			continue
		}
		switch split[0] {
		case "CLASS":
			notch, yarn := java.PackClassInfo(split[1]), java.PackClassInfo(split[3])
			mapping[notch] = yarn
			break
		case "METHOD":
			yarnClass, ok := mapping[java.PackClassInfo(split[1])]
			if !ok {
				continue
			}
			notch, yarn := java.PackMethodInfo(split[3], split[1], split[2]), java.PackMethodInfo(split[5], yarnClass.Signature, "")
			mapping[notch] = yarn
			break
		case "FIELD":
			yarnClass, ok := mapping[java.PackClassInfo(split[1])]
			if !ok {
				continue
			}
			notch, yarn := java.PackFieldInfo(split[3], split[1], split[2]), java.PackFieldInfo(split[5], yarnClass.Signature, "")
			mapping[notch] = yarn
			break
		}
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

func getMappingsTinyFromGzip(gzipData []byte) ([]byte, error) {
	gzipReader, err := gzip.NewReader(bytes.NewReader(gzipData))
	if err != nil {
		return nil, err
	}
	defer func(gzipReader *gzip.Reader) {
		err := gzipReader.Close()
		if err != nil {
			slog.Error("Error closing gzip reader: " + err.Error())
		}
	}(gzipReader)
	content, err := io.ReadAll(gzipReader)
	if err != nil {
		return nil, err
	}
	return content, nil
}
