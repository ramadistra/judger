package main

import (
	"bytes"
	"fmt"
	"io"
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
	port      string = ":8000"
	uploadDir string = ""
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

func main() {
	fmt.Println("Server running at http://localhost" + port)
	router := httprouter.New()
	router.HandlerFunc("POST", "/python3", python3)
	n := negroni.Classic()
	n.UseHandler(router)
	http.ListenAndServe(port, n)
}

func python3(w http.ResponseWriter, r *http.Request) {
	var (
		status = http.StatusInternalServerError
		err    error
		id     string
	)
	defer func() {
		if id != "" {
			exec.Command("rm", "-rf", id).Run()
		}
		if err != nil {
			http.Error(w, err.Error(), status)
		}
	}()

	id = generateFileName()
	CopyDir("../python3", id)

	source := r.FormValue("source")
	sourcefile, err := os.Create(id + "/source.py")
	if err != nil {
		return
	}
	sourcefile.WriteString(source)
	sourcefile.Close()

	input := r.FormValue("stdin") + "\n"
	inputfile, err := os.Create(id + "/input.txt")
	if err != nil {
		return
	}
	inputfile.WriteString(input)
	sourcefile.Close()

	build := exec.Command("docker", "build", "-t", id, id)
	err = build.Run()
	if err != nil {
		return
	}

	// Run the source code.
	cmd := exec.Command("docker", "run", id)
	timeout := getTimeOut(r.FormValue("timeout"))
	output, err := runExecutable(cmd, input, timeout+1000)
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

func runExecutable(cmd *exec.Cmd, input string, timeout int) (*Output, error) {

	// Set cmd stdout and stderr to bytes buffer.
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	// Get stdin writer.
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	// Run the executable.
	if err = cmd.Start(); err != nil {
		return nil, err
	}
	// Write to stdin.
	_, err = io.WriteString(stdin, input)
	if err != nil {
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
