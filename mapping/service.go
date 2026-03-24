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
	"strconv"
	"strings"
	"time"
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

func GenerateSource(mcVersion, mappingType string) (string, error) {
	start := time.Now()
	if !CanAddTask(mcVersion, mappingType) {
		return "", errors.New("this type has generated or generating")
	}
	service, ok := serviceMap[mappingType]
	if !ok {
		return "", errors.New("unknown mapping type")
	}

	slog.Info(fmt.Sprintf("Decompiling source type %s for %s", mappingType, mcVersion))
	StartPending(mcVersion, mappingType)
	path, err := service.Remap(mcVersion)
	if err != nil {
		FailurePending(mcVersion, mappingType)
		return "", err
	}
	sourcePath := global.GetSourceFolder(service, mcVersion)
	params := util.ConcatMultipleSlices([][]string{global.Config.Decompiler.JavaParams, {"-jar", global.DecompilerPath}, global.Config.Decompiler.DecompilerParams, {path, sourcePath}})
	err = util.ExecuteCommand(global.Config.JavaPath, params, true)
	if err != nil {
		FailurePending(mcVersion, mappingType)
		return "", err
	}
	Done(mcVersion, mappingType)
	slog.Info("Done in " + strconv.FormatInt(int64(time.Since(start)/1000000), 10) + "ms")
	return sourcePath, nil
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

	if IsPending(mcVersion, mappingType) {
		return "", errors.New("this mapping is pending")
	}

	jarPath, err := ensureRemappedJar(service, mcVersion)
	if err != nil {
		return "", err
	}

	classFile := filepath.Join("temp", className+".class")
	if err := os.MkdirAll(filepath.Dir(classFile), os.ModePerm); err != nil {
		return "", err
	}

	if err := util.ExtractClassFromJar(jarPath, className, classFile); err != nil {
		return "", err
	}

	params := util.ConcatMultipleSlices([][]string{global.Config.Decompiler.JavaParams, {"-jar", global.DecompilerPath}, global.Config.Decompiler.DecompilerParams, {classFile, filepath.Dir(targetPath)}})
	if err := util.ExecuteCommand(global.Config.JavaPath, params, true); err != nil {
		return "", err
	}

	return targetPath, nil
}
