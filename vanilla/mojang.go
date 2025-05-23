package vanilla

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"pluto/global"
	"pluto/util/network"
	"strings"
)

type SingleManifest struct {
	Id  string `json:"id"`
	Url string `json:"url"`
}

type VersionManifest struct {
	Versions []SingleManifest `json:"versions"`
}

type SingleFile struct {
	Sha1 string `json:"sha1"`
	Size int64  `json:"size"`
	Url  string `json:"url"`
}

type Downloads struct {
	Client         SingleFile     `json:"client"`
	ClientMappings SingleManifest `json:"client_mappings"`
}

type PistonData struct {
	Downloads Downloads `json:"downloads"`
}

var cache = map[string]Downloads{}

func GetOrDownload(mcVersion string) (Downloads, error) {
	if downloads, ok := cache[mcVersion]; ok {
		return downloads, nil
	}
	//request launcher meta
	data, err := network.Get(global.Config.Urls.MojangLauncherMeta + "/mc/game/version_manifest_v2.json")
	if err != nil {
		return Downloads{}, err
	}
	manifest := VersionManifest{}
	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return Downloads{}, err
	}
	var url = ""
	for _, version := range manifest.Versions {
		if version.Id == mcVersion {
			url = version.Url
			break
		}
	}
	if url == "" {
		return Downloads{}, errors.New("Cannot find mc version " + mcVersion)
	}
	//request piston data
	data, err = network.Get(replaceUrl(url))
	if err != nil {
		return Downloads{}, err
	}
	downloads := PistonData{}
	err = json.Unmarshal(data, &downloads)
	if err != nil {
		return Downloads{}, err
	}
	downloads.Downloads.Client.Url = replaceUrl(downloads.Downloads.Client.Url)
	downloads.Downloads.ClientMappings.Url = replaceUrl(downloads.Downloads.ClientMappings.Url)
	cache[mcVersion] = downloads.Downloads
	return downloads.Downloads, nil
}

func GetMcJarPath(mcVersion string) (string, error) {
	path := global.GetMinecraftPath(mcVersion)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return path, nil
	}
	downloads, err := GetOrDownload(mcVersion)
	if err != nil {
		slog.Error("Unable to download " + mcVersion + " meta : " + err.Error())
		return "", err
	}
	err = network.File(downloads.Client.Url, path)
	if err != nil {
		slog.Error("Unable to download " + mcVersion + " file : " + err.Error())
		return "", err
	}
	return path, nil
}

func replaceUrl(url string) string {
	return strings.ReplaceAll(strings.ReplaceAll(url, "https://piston-meta.mojang.com", global.Config.Urls.MojangPistonMeta), "https://piston-data.mojang.com", global.Config.Urls.MojangPistonData)
}
