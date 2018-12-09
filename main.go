package main

import "vary/prom_rest_exporter/jq"

func main() {

	jqInst := jq.New()
	defer jqInst.Close()

	jqInst.CompileProgram(".[] | select(.foo % 2 == 0)")

	results := jqInst.ProcessInput(
		`[
			{"foo": 7, "bar": "helloooo"},
			{"foo": 8, "bar": "world2"},
			{"foo": 9, "bar": "wadup"},
			{"foo": 10, "bar": "heheehe"}
		]`)

	for _, r := range results {
		r.PrettyPrint()
	}
}
