package main

import (
	"errors"
	"fmt"
	"github.com/alexflint/go-arg"
	"log"
	"os"
	"sort"
	"strings"
)

const (
	TOOLS_DIR_NAME = "tools"
	TOOLS_URL      = "https://github.com/tuxotron/docker-image-generator/releases/download/tools/tools.zip"
	APP_NAME       = "doig"
)

type commands struct {
	List []string
}

func getCommandList(tools []string, categories []string, toolsDb map[string]*Config) (map[string]*Config, error) {

	toolSet := make(map[string]*Config)

	for _, category := range categories {
		for k, v := range toolsDb {
			if category == "all" || category == v.Default.Category {
				toolSet[k] = v
			}
		}
	}

	for _, tool := range tools {
		if val, ok := toolsDb[tool]; ok {
			if _, ok := toolSet[tool]; !ok { // Check if the tool has been already added by a metapackage
				toolSet[tool] = val
			}
		} else {
			return nil, errors.New("[x] " + tool + " is not in the available tools")
		}
	}

	return toolSet, nil
}

type args struct {
	Tools      []string `arg:"-t" help:"List of tools separated by blank spaces"`
	Category   []string `arg:"-c" help:"List of categories separated by blank spaces"`
	Image      string   `arg:"-i" help:"Image name in lowercase"`
	Dockerfile bool     `arg:"-d" help:"Prints out the Dockerfile "`
	List       bool     `arg:"-l" help:"List the available tools and categories"`
	Update     bool     `arg:"-u" help:"Update tools"`
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

func dirDoesNotExists(dir string) bool {
	_, err := os.Stat(dir)
	return os.IsNotExist(err)
}

func setup(dir string) error {

	fmt.Println(Green("[*] Setting up doig ..."))
	err := os.MkdirAll(dir+"/"+TOOLS_DIR_NAME, 0755)
	if err != nil {
		fmt.Println(Red("[X] Error setting up the tools ..."))
		return err
	}

	err = UpdateTools(dir)
	if err != nil {
		return err
	}
	fmt.Println(Green("[*] doig setup complete ..."))

	return nil
}

func main() {

	var args args
	toolsDb := make(map[string]*Config)
	appDir := getAppDirectory()

	if dirDoesNotExists(appDir) {
		if err := setup(appDir); err != nil {
			log.Fatal(Red(err))
			os.Exit(1)
		}
	}

	parser := arg.MustParse(&args)
	if args.Update {
		fmt.Println(Green("[*] Updating tools ..."))
		err := UpdateTools(appDir)
		if err != nil {
			fmt.Println(Red("[X] Updating tools FAILED..."))
			log.Fatal(Red(err))
		}
		fmt.Println(Green("[*] Tools updated"))
	}

	loadTools(toolsDb, appDir+"/"+TOOLS_DIR_NAME)

	toolSet, err := getCommandList(args.Tools, args.Category, toolsDb)
	if err != nil {
		fmt.Println(Red(err))
		os.Exit(1)
	}

	if args.Dockerfile {
		if dockerfile, err := generateDockerfile(toolSet, appDir); err != nil {
			panic(err)
		} else {
			fmt.Println("\n" + dockerfile)
		}
		os.Exit(0)
	}

	if len(args.Image) > 0 {
		dockerfile, err := generateDockerfile(toolSet, appDir)
		if err != nil {
			log.Fatal(Red(err))
		}

		if len(toolSet) > 0 {
			// Copy into the image the file with the tools included in the image
			dockerfile = dockerfile + "\nCOPY tools.txt ."
		}

		dockerContext, err := createDockerContext([]byte(dockerfile), toolSet)
		if err != nil {
			log.Fatal(Red(err))
		}

		err = createDockerImage(dockerContext, strings.ToLower(args.Image))
		if err != nil {
			log.Fatal(Red(err))
		}

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
