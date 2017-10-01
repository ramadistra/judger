package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

var (
	// ImageDir is the location of the Docker images source file.
	ImageDir = "./images/"
	// DefaultTimeOut is the default timeout in milliseconds
	DefaultTimeOut = 1000
	// MaxTimeOut is the default maximum timeout in milliseconds
	MaxTimeOut = 10000
)

// Output contains the stdout, stderr and status of an execution.
type Output struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Status string `json:"status"`
}

// Input contains the necessary data for execution.
type Input struct {
	Source  string   `json:"source"`
	Stdin   []string `json:"stdin"`
	TimeOut int      `json:"timeout"`
}

// Image is a Docker image derived from a base image.
type Image struct {
	ID            string
	Ext           string
	BaseImagePath string
	Input         *Input
}

// Build builds the Docker Image from the Base Image.
func (image *Image) Build() error {
	// Copy base image and delete after build.
	CopyDir(ImageDir+image.BaseImagePath, image.ID)
	defer func() {
		exec.Command("rm", "-rf", image.ID).Run()
	}()

	// Write the source files to the new Image
	source := image.Input.Source
	sourcefile, err := os.Create(image.ID + "/source" + image.Ext)
	if err != nil {
		return err
	}
	sourcefile.WriteString(source)
	sourcefile.Close()

	// Write the inputs to the new Image.
	for i, v := range image.Input.Stdin {
		filename := fmt.Sprintf("/%d.in", i+1)
		inputfile, err := os.Create(image.ID + filename)
		if err != nil {
			return err
		}
		inputfile.WriteString(v + "\n")
		inputfile.Close()
	}

	build := exec.Command("docker", "build", "-t", image.ID, image.ID)
	if err = build.Run(); err != nil {
		return err
	}

	return nil
}

// Run runs the Docker Image and returns the output.
func (image *Image) Run() (*Output, error) {
	cmd := exec.Command("docker", "run", image.ID)
	return RunWithTimeOut(cmd, image.Input.TimeOut+1000)
}

// Remove deletes the Docker Image from Docker.
func (image *Image) Remove() error {
	return exec.Command("docker", "rmi", image.ID).Run()
}

// NewImage creates a new image and gives it and ID.
func NewImage(imagePath string, ext string, input *Input) *Image {
	id := generateID()
	return &Image{
		ID:            id,
		Input:         input,
		BaseImagePath: imagePath,
		Ext:           ext,
	}
}

// RunWithTimeOut runs an command with a set timeout.
func RunWithTimeOut(cmd *exec.Cmd, timeout int) (*Output, error) {

	// Set cmd stdout and stderr to bytes buffer.
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	// Run the executable.
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// Kill process if time limit is reached.
	var status string
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case <-time.After(time.Duration(timeout) * time.Millisecond):
		cmd.Process.Kill()
		status = "Timed out"
	case err := <-done:
		if err != nil {
			status = "Error"
		} else {
			status = "OK"
		}
	}

	return &Output{outbuf.String(), errbuf.String(), status}, nil
}

func generateID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

func getTimeOut(inputTimeOut string) int {
	tempTimeOut, err := strconv.Atoi(inputTimeOut)
	if err != nil {
		return DefaultTimeOut
	}
	if tempTimeOut < MaxTimeOut {
		return tempTimeOut
	}
	return MaxTimeOut
}
