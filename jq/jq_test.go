package jq

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRunJqProgram(t *testing.T) {
	jqInst := New()
	defer jqInst.Close()

	jqInst.CompileProgram(".[] | select(.foo % 2 == 0) | .bar")

	results := jqInst.ProcessInput(
		`[
			{"foo": 7, "bar": "helloooo"},
			{"foo": 8, "bar": "world"},
			{"foo": 9, "bar": "wadup"},
			{"foo": 10, "bar": "heheehe"}
		]`)

	expected := [...]string{"world", "heheehe"}
	for i := range results {
		assert.Equal(t, expected[i], results[i].ToString())
	}
}
