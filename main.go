package main

import (
	"flag"
	"log"
	"log/slog"
	"os"
	"pluto/global"
	"pluto/mapping"
	"pluto/util"
	"pluto/webserver"
)

func main() {
	//Params
	skipLibraryCheck := flag.Bool("skiplibrary", false, "Skip Library Check")
	flag.Parse()
	//Main
	defer util.CloseWorkers()
	util.InitLogger()
	slog.Info("Launching Pluto v" + global.Version)
	//Config
	slog.Info("Loading configs...")
	err := global.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	err = mapping.InitMappingConfig()
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll("temp", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	//Libraries
	if !*skipLibraryCheck {
		global.CheckLibrary()
	}
	//Main Logic
	err = webserver.Launch()
	if err != nil {
		log.Fatal(err)
	}
}
