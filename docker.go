package main

import (
	"archive/tar"
	"bytes"
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"log"
	"os"
	"text/template"
)

func generateDockerfile(toolSet map[string]*Config, appDir string) (string, error) {

	cmds := commands{}
	for _, v := range toolSet {
		cmds.List = append(cmds.List, v.Default.Command)
	}

	t := template.New("Dockerfile.template")
	t, err := t.ParseFiles(appDir + "/Dockerfile.template")
	if err != nil {
		log.Fatal(err)
	}

	var tpl bytes.Buffer
	err = t.Execute(&tpl, cmds)
	if err != nil {
		log.Fatal(err)
	}

	return tpl.String(), nil
}

func createDockerContext(dockerfile []byte, toolSet map[string]*Config) (*bytes.Reader, error) {

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
		for k := range toolSet {
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

func createDockerImage(dockerContext *bytes.Reader, imageName string) error {

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
	err = jsonmessage.DisplayJSONMessagesStream(imageBuildResponse.Body, os.Stdout, os.Stdout.Fd(), true, nil)

	return err

}
