package services

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"pluto/global"
	"pluto/mapping/java"
	"pluto/util"
	"pluto/util/network"
	"pluto/vanilla"
	"strings"
)

type Official struct{}

var officialMappings = make(map[string]*java.Mappings)

func (s *Official) GetName() string {
	return "official"
}

func (s *Official) GetPathOrDownload(mcVersion string) (string, error) {
	path := global.GetMappingPath(s, mcVersion, "txt")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return path, nil
	}
	downloads, err := vanilla.GetOrDownload(mcVersion)
	if err != nil {
		return "", err
	}
	data, err := network.Get(downloads.ClientMappings.Url)
	if err != nil {
		return "", err
	}
	err = os.WriteFile(path, data, 0666)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (s *Official) GetMappingCacheOrError(mcVersion string) (*java.Mappings, error) {
	if mapping, ok := officialMappings[mcVersion]; ok {
		return mapping, nil
	}
	return nil, errors.New("not cached yet")
}

func (s *Official) SaveMappingCache(mcVersion string, mapping *java.Mappings) {
	officialMappings[mcVersion] = mapping
}

func (s *Official) LoadMapping(mcVersion string) (*map[java.SingleInfo]java.SingleInfo, error) {
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
	cachedNotchClass, cachedNamedClass := java.SingleInfo{}, java.SingleInfo{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || len(line) == 0 {
			continue
		}
		split := strings.Split(line, " -> ")
		if len(split) != 2 {
			continue
		}
		if !strings.HasPrefix(split[0], "    ") { //class
			notch, named := java.PackClassInfo(strings.ReplaceAll(split[1], ":", "")), java.PackClassInfo(split[0])
			mapping[notch] = named
			cachedNotchClass, cachedNamedClass = notch, named
		} else if strings.Contains(split[0], ":") { //method
			s := strings.Split(split[0], ":")
			if len(s) == 3 {
				name := strings.Split(strings.Split(s[2], " ")[1], "(")[0]
				signature, err := java.MethodToByteCodeSignature(s[2], false)
				if err != nil {
					continue
				}
				notch, named := java.PackMethodInfo(split[1], cachedNotchClass.Class, ""), java.PackMethodInfo(name, cachedNamedClass.Class, signature)
				mapping[notch] = named
			}
		} else { //Field
			s := strings.Split(strings.TrimSpace(split[0]), " ")
			if len(s) == 2 {
				notch, named := java.PackFieldInfo(split[1], cachedNotchClass.Class, ""), java.PackFieldInfo(s[1], cachedNamedClass.Class, java.ClassToByteCodeSignature(s[0]))
				mapping[notch] = named
			}
		}
	}
	//Post processor
	result, classMapping := make(map[java.SingleInfo]java.SingleInfo), make(map[string]string)
	for notch, named := range mapping {
		if notch.Type == "class" {
			classMapping[named.Signature] = notch.Signature
		}
	}
	for notch, named := range mapping {
		notch.Name = java.FullToClassName(notch.Name)
		named.Name = java.FullToClassName(named.Name)
		switch notch.Type {
		case "method":
			notch.Signature = java.ObfuscateMethodSignature(named.Signature, classMapping)
		case "field":
			notch.Signature = java.ObfuscateTypeSignature(named.Signature, classMapping)
		}
		result[notch] = named
	}
	return &result, nil
}

func (s *Official) Remap(mcVersion string) (string, error) {
	jarPath, err := vanilla.GetMcJarPath(mcVersion)
	if err != nil {
		return "", err
	}
	mappingPath, err := s.GetPathOrDownload(mcVersion)
	if err != nil {
		return "", err
	}
	outputPath := global.GetRemappedPath(s, mcVersion)
	err = util.ExecuteCommand(global.Config.JavaPath, []string{"-cp", global.ClassPath, global.ArtMainClass, "--input", jarPath, "--output", outputPath, "--map", mappingPath, "--reverse"}, true)
	if err != nil {
		return "", err
	}
	return outputPath, nil
}
