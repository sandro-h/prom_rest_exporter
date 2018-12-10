package main

import (
	"io/ioutil"
	"net/http"
	"vary/prom_rest_exporter/jq"
)

func main() {
	jqInst := jq.New()
	defer jqInst.Close()

	jqInst.CompileProgram("[.data[].last_name] | length")

	input, _ := Fetch("https://reqres.in/api/users")

	results := jqInst.ProcessInput(input)

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
