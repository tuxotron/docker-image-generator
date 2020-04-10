package main

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/spf13/viper"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

type config struct {
	Default struct {
		Name     string `mapstructure:"name"`
		Command  string `mapstructure:"command"`
		Category string `mapstructure:"category"`
	}
}

func filenameWithoutExtension(fn string) string {
	return strings.TrimSuffix(fn, path.Ext(fn))
}

func loadTools(toolsDb map[string]*config) {
	myviper := viper.New()
	//myviper.SetConfigType("ini")   // REQUIRED if the config file does not have the extension in the name

	directory := "tools"
	myviper.AddConfigPath(directory)

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {

		if filepath.Ext(path) == ".ini" {
			myviper.SetConfigName(filenameWithoutExtension(info.Name()))

			err = myviper.ReadInConfig()
			if err != nil {
				panic(fmt.Errorf("Fatal error config file: %s \n", err))
			}
			cfg := new(config)
			err = myviper.Unmarshal(cfg)
			if err != nil {
				panic(fmt.Errorf("Fatal error unmarshaling config file: %s \n", err))
			}

			toolsDb[myviper.GetString("default.name")] = cfg
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func createDockerImage(dockerContext *bytes.Reader, imageName string) {

	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	imageBuildResponse, err := cli.ImageBuild(ctx, dockerContext,
		types.ImageBuildOptions{
			Tags: []string{imageName},
		})
	if err != nil {
		panic(err)
	}

	defer imageBuildResponse.Body.Close()
	_, err = io.Copy(os.Stdout, imageBuildResponse.Body)
	if err != nil {
		log.Fatal(err, " :unable to read image build response")
	}

	//reader, err := cli.ImagePull(ctx, "docker.io/library/ubuntu:18.04", types.ImagePullOptions{})
	//if err != nil {
	//	panic(err)
	//}
	//io.Copy(os.Stdout, reader)

}

type commands struct {
	List []string
}

func generateDockerfile(cmds *commands) (string, error) {

	t := template.New("Dockerfile.template")
	t, err := t.ParseFiles("Dockerfile.template")
	if err != nil {
		panic(err)
	}

	var tpl bytes.Buffer
	err = t.Execute(&tpl, cmds)
	if err != nil {
		panic(err)
	}

	return tpl.String(), nil
}

func getCommandList(tools []string, metapackages []string, toolsDb map[string]*config) commands {

	cmds := commands{}

	for _, tool := range tools {
		if val, ok := toolsDb[tool]; ok {
			l := append(cmds.List, val.Default.Command)
			cmds.List = l
		} else {
			fmt.Println("[*] " + tool + " is not in the available tools")
		}
	}

	return cmds
}

func createDockerContext(dockerfile []byte) (*bytes.Reader, error) {

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	tarHeader := &tar.Header{
		Name: "Dockerfile",
		Size: int64(len(dockerfile)),
	}
	err := tw.WriteHeader(tarHeader)
	if err != nil {
		log.Fatal(err, " :unable to write tar header")
		return nil, err
	}
	_, err = tw.Write(dockerfile)
	if err != nil {
		log.Fatal(err, " :unable to write tar body")
		return nil, err
	}
	return bytes.NewReader(buf.Bytes()), nil

}

type args struct {
	Tools        []string `arg:"-t" help:"List of tools separated by blank spaces"`
	Metapackages []string `arg:"-m" help:"List of metapackages separated by blank spaces"`
	Image        string   `arg:"-i" help:"Image name"`
	Dockerfile   bool     `arg:"-d" help:"Prints out the Dockerfile "`
	List         bool     `arg:"-l" help:"List the available tools"`
}

func (args) Description() string {
	return "This tool creates a customized docker image with the tools you need\n"
}

func main() {

	var args args
	toolsDb := make(map[string]*config)

	parser := arg.MustParse(&args)
	loadTools(toolsDb)
	cmds := getCommandList(args.Tools, args.Metapackages, toolsDb)

	if args.Dockerfile {
		if dockerfile, err := generateDockerfile(&cmds); err != nil {
			panic(err)
		} else {
			fmt.Println("\n" + dockerfile)
		}
		os.Exit(0)
	}

	if len(args.Image) > 0 {
		dockerfile, err := generateDockerfile(&cmds)
		if err != nil {
			panic(err)
		}

		dockerContext, err := createDockerContext([]byte(dockerfile))
		if err != nil {
			panic(err)
		}

		createDockerImage(dockerContext, strings.ToLower(args.Image))
		os.Exit(0)
	}

	if args.List {
		keys := make([]string, 0, len(toolsDb))
		for k, _ := range toolsDb {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("%s\n", k)
		}

		os.Exit(0)
	}

 	parser.WriteHelp(os.Stdout)

}
