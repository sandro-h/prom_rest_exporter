package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"vary/prom_rest_exporter/jq"
)

func main() {
	jqInst := jq.New()
	defer jqInst.Close()

	err := jqInst.CompileProgram("[.data[].last_name] | length")
	if err != nil {
		fmt.Printf("Compile error: %s", err)
		return
	}

	input, _ := Fetch("https://reqres.in/api/users")

	results, err := jqInst.ProcessInput(input)
	if err != nil {
		fmt.Printf("Process error: %s", err)
		return
	}

	for _, r := range results {
		r.PrettyPrint()
	}
}

// Fetch calls the url and returns the response as a string
func Fetch(url string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
