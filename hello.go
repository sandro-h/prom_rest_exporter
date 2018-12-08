package main

/*
#cgo CFLAGS: -Ijq-master/build/win64/usr/local/include
#cgo LDFLAGS: -Ljq-master/build/win64/usr/local/lib -ljq -lonig -lshlwapi
#cgo linux LDFLAGS: -lm
#include <jq.h>
#include <jv.h>
#include <stdlib.h>
*/
import "C"
import "fmt"

func main() {

	state := C.jq_init()

	// Compile jq program
	prog := C.CString(".[] | select(.foo % 2 == 0) | .bar")
	args := C.jv_array()
	C.jq_compile_args(state, prog, args)

	// Prepare input
	flags := C.int(0)
	input := `
	[
		{"foo": 7, "bar": "helloooo"},
		{"foo": 8, "bar": "world"},
		{"foo": 9, "bar": "wadup"},
		{"foo": 10, "bar": "heheehe"}
	]
	`
	jvInput := C.jv_parse(C.CString(input))

	// Process input
	dumpFlags := C.int(C.JV_PRINT_PRETTY | C.JV_PRINT_SPACE1)
	C.jq_start(state, jvInput, flags)
	res := C.jq_next(state)
	for C.jv_is_valid(res) != 0 {
		resStr := C.jv_dump_string(res, dumpFlags)
		fmt.Printf("%s\n", C.GoString(C.jv_string_value(resStr)))
		res = C.jq_next(state)
	}
}

func foo(x int) int {
	return 2 * x
}
