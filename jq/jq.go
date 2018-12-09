package jq

/*
#cgo CFLAGS: -I../jq-master/build/win64/usr/local/include
#cgo LDFLAGS: -L../jq-master/build/win64/usr/local/lib -ljq -lonig -lshlwapi
#cgo linux LDFLAGS: -lm
#include <jq.h>
#include <jv.h>
#include <stdlib.h>
*/
import "C"
import "unsafe"
import "fmt"

// Jq represents a a Jq state
type Jq struct {
	state *C.struct_jq_state
}

// New creates a new instance of Jq containing the jq state
func New() *Jq {
	jq := new(Jq)
	jq.state = C.jq_init()
	return jq
}

// Close frees the resources associated with this Jq instance
func (jq *Jq) Close() {
	C.jq_teardown(&jq.state)
	jq.state = nil
}

// CompileProgram compiles a jq program for the passed Jq instance
func (jq *Jq) CompileProgram(prog string) {
	csProg := C.CString(prog)
	defer C.free(unsafe.Pointer(csProg))
	args := C.jv_array()
	C.jq_compile_args(jq.state, csProg, args)
}

// ProcessInput runs the previously compiled program of the Jq instance on the input
func (jq *Jq) ProcessInput(input string) []*Jv {
	results := make([]*Jv, 0)
	flags := C.int(0)

	csInput := C.CString(input)
	defer C.free(unsafe.Pointer(csInput))
	jvInput := C.jv_parse(csInput)
	defer C.jv_free(jvInput)

	C.jq_start(jq.state, jvInput, flags)
	res := C.jq_next(jq.state)
	for C.jv_is_valid(res) != 0 {
		results = append(results, &Jv{res})
		res = C.jq_next(jq.state)
	}

	return results
}

// Jv represents a Jq json value (jv)
type Jv struct {
	jv C.jv
}

func (jv *Jv) _ToString(flags C.int) string {
	jvStr := C.jv_dump_string(jv.jv, flags)
	defer C.jv_free(jvStr)
	str := C.GoString(C.jv_string_value(jvStr))
	// Always use "raw" output: no quotes around strings
	if C.jv_get_kind(jv.jv) == C.JV_KIND_STRING {
		str = str[1 : len(str)-1]
	}
	return str
}

// ToString returns a non-pretty-print string representation of the json value
func (jv *Jv) ToString() string {
	return jv._ToString(C.int(0))
}

// PrettyPrint pretty prints the json value to stdout
func (jv *Jv) PrettyPrint() {
	prettyFlags := C.int(C.JV_PRINT_PRETTY | C.JV_PRINT_SPACE1)
	fmt.Printf("%s\n", jv._ToString(prettyFlags))
}
