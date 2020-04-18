package main

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Config struct {
	Default struct {
		Name     string `mapstructure:"name"`
		Command  string `mapstructure:"command"`
		Category string `mapstructure:"category"`
		Comment  string `mapstructure:"comment"`
	}
}

func filenameWithoutExtension(fn string) string {
	return strings.TrimSuffix(fn, path.Ext(fn))
}

func getAppDirectory() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	return home + "/." + APP_NAME //+ "/" + TOOLS_DIR_NAME
}

func loadTools(toolsDb map[string]*Config, directory string) {
	myviper := viper.New()
	myviper.AddConfigPath(directory)

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {

		if filepath.Ext(path) == ".ini" {
			myviper.SetConfigName(filenameWithoutExtension(info.Name()))

			err = myviper.ReadInConfig()
			if err != nil {
				panic(fmt.Errorf("Fatal error config file: %s \n", err))
			}
			cfg := new(Config)
			err = myviper.Unmarshal(cfg)
			if err != nil {
				panic(fmt.Errorf("Fatal error unmarshaling config file: %s \n", err))
			}

			toolsDb[myviper.GetString("default.name")] = cfg
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
