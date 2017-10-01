package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/json-iterator/go"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
)

const port string = ":8000"

var json = jsoniter.ConfigCompatibleWithStandardLibrary

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
			err   error
			image *Image
		)
		defer func() {
			if err != nil {
				http.Error(w, err.Error(), 500)
			}
			image.Remove()
		}()

		// Parse incoming request.
		var input Input
		inputJSON, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return
		}
		err = json.Unmarshal(inputJSON, &input)
		if err != nil {
			return
		}

		// Build a new docker image.
		image = NewImage(imagePath, ext, &input)
		if err = image.Build(); err != nil {
			return
		}
		// Run the image.
		output, err := image.Run()
		if err != nil {
			return
		}

		// Encode the output to JSON.
		respJSON, err := json.Marshal(output)
		if err != nil {
			return
		}

		w.Write(respJSON)
	}
}
