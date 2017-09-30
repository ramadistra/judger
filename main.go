package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/json-iterator/go"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
)

const (
	port     string = ":8000"
	imageDir string = "./images/"
	// Milliseconds
	defaultTimeOut int = 1000
	maxTimeOut     int = 10000
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Output contains the stdout, stderr and status of an execution.
type Output struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Status string `json:"status"`
}

type APIRequest struct {
	Source  string   `json:"source"`
	Stdin   []string `json:"stdin"`
	TimeOut int      `json:"timeout"`
}

func main() {
	fmt.Println("Server running at http://localhost" + port)
	router := httprouter.New()
	router.HandlerFunc("POST", "/python3", handleImage("python3", ".py"))
	n := negroni.Classic()
	n.UseHandler(router)
	http.ListenAndServe(port, n)
}

func handleImage(imagePath string, ext string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			status = http.StatusInternalServerError
			err    error
			id     string
		)
		defer func() {
			if id != "" {
				// Delete folder.
				exec.Command("rm", "-rf", id).Run()
				// Delete image.
				exec.Command("docker", "rmi", id).Run()
			}
			if err != nil {
				http.Error(w, err.Error(), status)
			}
		}()

		var req APIRequest
		reqJSON, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return
		}
		err = json.Unmarshal(reqJSON, &req)
		if err != nil {
			return
		}

		id = generateFileName()

		// Copy base image
		CopyDir(imageDir+imagePath, id)

		source := req.Source
		sourcefile, err := os.Create(id + "/source.py")
		if err != nil {
			return
		}
		sourcefile.WriteString(source)
		sourcefile.Close()

		for i, v := range req.Stdin {
			filename := fmt.Sprintf("/%d.in", i+1)
			inputfile, err := os.Create(id + filename)
			if err != nil {
				return
			}
			inputfile.WriteString(v + "\n")
			inputfile.Close()
		}

		build := exec.Command("docker", "build", "-t", id, id)
		err = build.Run()
		if err != nil {
			return
		}

		// Run the source code.
		cmd := exec.Command("docker", "run", id)
		timeout := getTimeOut(r.FormValue("timeout"))
		output, err := runExecutable(cmd, timeout+1000)
		if err != nil {
			return
		}

		// Encode output to JSON.
		respJSON, err := json.Marshal(output)
		if err != nil {
			return
		}

		w.Write(respJSON)
	}
}

func runExecutable(cmd *exec.Cmd, timeout int) (*Output, error) {

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

func generateFileName() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

func getTimeOut(inputTimeOut string) int {
	tempTimeOut, err := strconv.Atoi(inputTimeOut)
	if err != nil {
		return defaultTimeOut
	}
	if tempTimeOut < maxTimeOut {
		return tempTimeOut
	}
	return maxTimeOut
}
