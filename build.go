package main

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var dockerfile = `
FROM alpine:3.8
RUN echo "hello world!"
`

const imageName = "build:sample"

func main() {
	ctx := context.Background()

	f, err := os.CreateTemp("", "Dockerfile.*")
	if err != nil {
		log.Fatal(err)
	}
	dockerfileName := f.Name()
	defer func() {
		_ = f.Close()
		_ = os.Remove(dockerfileName)
	}()

	_, err = f.WriteString(dockerfile)
	if err != nil {
		panic(err)
	}

	cli, err := client.NewClientWithOpts()
	if err != nil {
		panic(err)
	}

	archived, err := archive(dockerfileName, dockerfile)
	if err != nil {
		panic(err)
	}

	resp, err := cli.ImageBuild(ctx, archived, types.ImageBuildOptions{
		Dockerfile: dockerfileName,
		Tags:       []string{imageName},
		Remove:     true,
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		panic(err)
	}
}

func archive(dockerfileName, dockerfile string) (io.Reader, error) {
	buf := &bytes.Buffer{}
	tw := tar.NewWriter(buf)

	err := tw.WriteHeader(&tar.Header{
		Name: dockerfileName,
		Size: int64(len(dockerfile)),
	})
	if err != nil {
		return nil, err
	}
	_, err = tw.Write([]byte(dockerfile))
	if err != nil {
		return nil, err
	}
	err = tw.Close()
	if err != nil {
		return nil, err
	}

	return buf, nil
}
