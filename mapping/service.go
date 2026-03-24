package mapping

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"pluto/global"
	"pluto/mapping/java"
	"pluto/mapping/services"
	"pluto/util"
	"strings"
)

type Service interface {
	GetName() string
	GetPathOrDownload(mcVersion string) (string, error)
	GetMappingCacheOrError(mcVersion string) (*java.Mappings, error)
	SaveMappingCache(mcVersion string, mapping *java.Mappings)
	LoadMapping(mcVersion string) (*map[java.SingleInfo]java.SingleInfo, error) //All default is notch->target
	Remap(mcVersion string) (string, error)
}

var (
	serviceMap = map[string]Service{
		"official": &services.Official{},
		"yarn":     &services.Yarn{},
	}
	loadMappingLock = util.NewNamedLock()
)

func LoadMapping(mcVersion, mappingType string) (*java.Mappings, error) {
	service, ok := serviceMap[mappingType]
	if !ok {
		return &java.Mappings{}, errors.New("unknown mapping type")
	}
	if cache, err := service.GetMappingCacheOrError(mcVersion); err == nil {
		return cache, nil
	}

	if loadMappingLock.IsLocked(mcVersion, mappingType) {
		slog.Warn("This mapping is loading!")
		return nil, errors.New("this mapping is loading")
	}
	loadMappingLock.Lock(mcVersion, mappingType)
	defer loadMappingLock.Unlock(mcVersion, mappingType)

	slog.Info(fmt.Sprintf("Loading mapping type %s for %s", mappingType, mcVersion))
	m, err := service.LoadMapping(mcVersion)
	if err != nil {
		return &java.Mappings{}, err
	}
	m3 := java.BuildMapping(m)
	service.SaveMappingCache(mcVersion, m3)
	return m3, nil
}

func ensureRemappedJar(service Service, mcVersion string) (string, error) {
	jarPath := global.GetRemappedPath(service, mcVersion)
	if _, err := os.Stat(jarPath); err == nil {
		return jarPath, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	return service.Remap(mcVersion)
}

func GenerateSourceForClass(mcVersion, mappingType, className string) (string, error) {
	service, ok := serviceMap[mappingType]
	if !ok {
		return "", errors.New("unknown mapping type")
	}

	className = strings.TrimSpace(className)
	if className == "" {
		return "", errors.New("class path is empty")
	}
	className = strings.ReplaceAll(className, ".", "/")
	className = strings.Trim(className, "/")

	sourceDir := global.GetSourceFolder(service, mcVersion)
	targetPath := filepath.Join(sourceDir, className+".java")
	if _, err := os.Stat(targetPath); err == nil {
		return targetPath, nil
	}
	targetFolder := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetFolder, os.ModePerm); err != nil {
		return "", err
	}

	jarPath, err := ensureRemappedJar(service, mcVersion)
	if err != nil {
		return "", err
	}

	classFiles, err := util.ExtractClassFromJar(jarPath, className, "temp")
	if err != nil {
		return "", err
	}

	params := util.ConcatMultipleSlices([][]string{global.Config.Decompiler.JavaParams, {"-jar", global.DecompilerPath}, global.Config.Decompiler.DecompilerParams, classFiles, {targetFolder}})
	if err := util.ExecuteCommand(global.Config.JavaPath, params, true); err != nil {
		return "", err
	}

	return targetPath, nil
}
