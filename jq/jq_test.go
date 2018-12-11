package jq

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRunJqProgram(t *testing.T) {
	jqInst := New()
	defer jqInst.Close()

	jqInst.CompileProgram(".[] | select(.foo % 2 == 0) | .bar")

	results, err := jqInst.ProcessInput(
		`[
			{"foo": 7, "bar": "helloooo"},
			{"foo": 8, "bar": "world"},
			{"foo": 9, "bar": "wadup"},
			{"foo": 10, "bar": "heheehe"}
		]`)

	assert.Nil(t, err)
	expected := [...]string{"world", "heheehe"}
	assert.Equal(t, len(expected), len(results), "should have same length")
	for i := range results {
		assert.Equal(t, expected[i], results[i].ToString())
	}
}

func TestCompileInvalidProgram(t *testing.T) {
	jqInst := New()
	defer jqInst.Close()
	err := jqInst.CompileProgram(".(]")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "jq: error: syntax error, unexpected")
}

func TestProcessInvalidInput(t *testing.T) {
	jqInst := New()
	defer jqInst.Close()

	jqInst.CompileProgram(".")

	_, err := jqInst.ProcessInput("[a;]")

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "jq: error: Invalid numeric literal at line")
}
