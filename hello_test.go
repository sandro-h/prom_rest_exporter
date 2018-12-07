package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFoo(t *testing.T) {
	res := foo(7)
	assert.Equal(t, 14, res)
}
