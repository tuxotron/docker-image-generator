package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"log"
	"net/http"
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
		Status   string `mapstructure:"status"`
	}
}

func filenameWithoutExtension(fn string) string {
	return strings.TrimSuffix(fn, path.Ext(fn))
}

func getAppDirectory() string {

	if len(os.Getenv("DOIG_PATH")) > 0 {
		return os.Getenv("DOIG_PATH")
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		return home + "/." + APP_NAME
	}

}

func loadTools(toolsDb map[string]*Config, directory string) {
	myviper := viper.New()
	myviper.AddConfigPath(directory)

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {

		if filepath.Ext(path) == ".ini" {
			myviper.SetConfigName(filenameWithoutExtension(info.Name()))

			err = myviper.ReadInConfig()
			if err != nil {
				log.Fatal(fmt.Errorf("Fatal error config file: %s \n", err))
			}
			cfg := new(Config)
			err = myviper.Unmarshal(cfg)
			if err != nil {
				log.Fatal(fmt.Errorf("Fatal error unmarshaling config file: %s \n", err))
			}

			if (cfg.Default.Status != "disabled") {
				toolsDb[myviper.GetString("default.name")] = cfg
			}

		}
		return nil
	})
	if err != nil {
		log.Fatal(Red(err))
	}
}

func DownloadFile(url string) ([]byte, error) {

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func Unzip(zipFile []byte, dir string) error {

	r, err := zip.NewReader(bytes.NewReader(zipFile), int64(len(zipFile)))
	if err != nil {
		return err
	}
	for _, zf := range r.File {
		if zf.FileInfo().IsDir() {
			err := os.MkdirAll(filepath.Join(dir, zf.Name), os.ModePerm)
			if err != nil {
				fmt.Println(Red("[X] Error creating " + filepath.Join(dir, zf.Name)))
				return err
			}
			continue
		}

		dst, err := os.Create(dir + "/" + zf.Name)
		if err != nil {
			fmt.Println(Red("[X] Error creating " + dir + "/" + zf.Name))
			return err
		}
		defer dst.Close()
		src, err := zf.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		_, err = io.Copy(dst, src)
		if err != nil {
			return nil
		}
	}

	return nil
}

func UpdateTools(dir string) error {
	fmt.Println(Green("[*] Updating tools ..."))
	zipFile, err := DownloadFile(TOOLS_URL)
	if err != nil {
		fmt.Println(Red("[X] Error downloading tools ..."))
		return err
	}

	err = Unzip(zipFile, dir)
	if err != nil {
		fmt.Println(Red("[X] Error unzipping tools ..."))
		return err
	}

	return nil
}
