package main

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/spf13/viper"
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
		Comment  string `mapstructure:"comment"`
	}
}

func filenameWithoutExtension(fn string) string {
	return strings.TrimSuffix(fn, path.Ext(fn))
}

func loadTools(toolsDb map[string]*config) {
	myviper := viper.New()
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
	jsonmessage.DisplayJSONMessagesStream(imageBuildResponse.Body, os.Stdout, os.Stdout.Fd(), true, nil)

}

type commands struct {
	List []string
}

func generateDockerfile(toolSet map[string]*config) (string, error) {

	cmds := commands{}
	for _, v := range toolSet {
		cmds.List = append(cmds.List, v.Default.Command)
	}

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

func getCommandList(tools []string, categories []string, toolsDb map[string]*config) map[string]*config {

	toolSet := make(map[string]*config)

	for _, category := range categories {
		for k, v := range toolsDb {
			if category == "all" || category == v.Default.Category {
				toolSet[k] = v
				//fmt.Println(Green("[*] Adding " + k))
			}
		}
	}

	for _, tool := range tools {
		if val, ok := toolsDb[tool]; ok {
			if _, ok := toolSet[tool]; !ok { // Check if the tool has been already added by a metapackage
				toolSet[tool] = val
				//fmt.Println(Green("[*] Adding " + tool))
			}
		} else {
			fmt.Println(Red("[x] " + tool + " is not in the available tools"))
			os.Exit(1)
		}
	}

	return toolSet
}

func createDockerContext(dockerfile []byte, toolSet map[string]*config) (*bytes.Reader, error) {

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	dockerfileHeader := &tar.Header{
		Name: "Dockerfile",
		Size: int64(len(dockerfile)),
	}
	err := tw.WriteHeader(dockerfileHeader)
	if err != nil {
		log.Fatal(err, " :unable to write dockerfile tar header")
		return nil, err
	}

	_, err = tw.Write(dockerfile)
	if err != nil {
		log.Fatal(err, " :unable to write tar dockerfile")
		return nil, err
	}

	if len(toolSet) > 0 {
		var toolList string
		for k, _ := range toolSet {
			toolList = toolList + k + "\n"
		}

		toolsHeader := &tar.Header{
			Name: "tools.txt",
			Size: int64(len(toolList)),
		}

		err = tw.WriteHeader(toolsHeader)
		if err != nil {
			log.Fatal(err, " :unable to write tools.txt tar header")
			return nil, err
		}

		_, err = tw.Write([]byte(toolList))
		if err != nil {
			log.Fatal(err, " :unable to write tar tools.txt")
			return nil, err
		}
	}

	return bytes.NewReader(buf.Bytes()), nil

}

type args struct {
	Tools      []string `arg:"-t" help:"List of tools separated by blank spaces"`
	Category   []string `arg:"-c" help:"List of categories separated by blank spaces"`
	Image      string   `arg:"-i" help:"Image name in lowercase"`
	Dockerfile bool     `arg:"-d" help:"Prints out the Dockerfile "`
	List       bool     `arg:"-l" help:"List the available tools and categories"`
}

func (args) Description() string {
	return "This tool creates a customized docker image with the tools you need\n"
}

var (
	Red    = Color("\033[1;31m%s\033[0m")
	Green  = Color("\033[1;32m%s\033[0m")
	Yellow = Color("\033[1;33m%s\033[0m")
)

func Color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}

func main() {

	var args args
	toolsDb := make(map[string]*config)

	parser := arg.MustParse(&args)
	loadTools(toolsDb)
	toolSet := getCommandList(args.Tools, args.Category, toolsDb)

	if args.Dockerfile {
		if dockerfile, err := generateDockerfile(toolSet); err != nil {
			panic(err)
		} else {
			fmt.Println("\n" + dockerfile)
		}
		os.Exit(0)
	}

	if len(args.Image) > 0 {
		dockerfile, err := generateDockerfile(toolSet)
		if err != nil {
			panic(err)
		}

		if len(toolSet) > 0 {
			// Copy into the image the file with the tools included in the image
			dockerfile = dockerfile + "\nCOPY tools.txt ."
		}

		dockerContext, err := createDockerContext([]byte(dockerfile), toolSet)
		if err != nil {
			panic(err)
		}

		createDockerImage(dockerContext, strings.ToLower(args.Image))

		fmt.Println(Green("\nTools added to the image:"))
		for _, v := range toolSet {
			if len(v.Default.Comment) > 0 {
				fmt.Println(Yellow("  [-] " + v.Default.Name + ": " + v.Default.Comment))
			} else {
				fmt.Println(Green("  [-] " + v.Default.Name))
			}
		}
		os.Exit(0)
	}

	if args.List {
		categories := make(map[string]struct{})
		keys := make([]string, 0, len(toolsDb))

		for k, v := range toolsDb {
			keys = append(keys, k)
			categories[v.Default.Category] = struct{}{}
		}
		sort.Strings(keys)

		fmt.Println(Green("[*] Tools"))
		for _, k := range keys {
			fmt.Println(Yellow("  [-] ", k))
		}

		fmt.Println(Green("[*] Categories"))
		for k, _ := range categories {
			fmt.Println(Yellow("  [-] ", k))
		}

		os.Exit(0)
	}
	parser.WriteHelp(os.Stdout)
}
