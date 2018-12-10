package jq

/*
#cgo windows CFLAGS: -I../jq-master/build/win64/usr/local/include
#cgo windows LDFLAGS: -L../jq-master/build/win64/usr/local/lib -ljq -lonig -lshlwapi
#cgo linux CFLAGS: -I../jq-master/build/linux/usr/local/include
#cgo linux LDFLAGS: -L../jq-master/build/linux/usr/local/lib -ljq -lonig -lm
#include <jq.h>
#include <jv.h>
#include <stdlib.h>

void install_jq_error_cb(jq_state *jq, unsigned long long id);
*/
import "C"
import "unsafe"
import "fmt"
import "errors"

// Jq represents a Jq state
type Jq struct {
	state          *C.struct_jq_state
	errorHandlerID uint64
	errorHandler   func(err string)
}

// Note: you can only set one error handler per jq instance

type JqErrorHandler struct {
	nextID    uint64
	callbacks map[uint64]func(err string)
}

var globalJqErrorHandler = JqErrorHandler{
	callbacks: make(map[uint64]func(err string)),
}

func (eh *JqErrorHandler) addErrorHandler(jq *Jq) {
	// TODO: not threadsafe
	jq.errorHandlerID = eh.nextID
	eh.callbacks[eh.nextID] = jq.handleError
	C.install_jq_error_cb(jq.state, C.ulonglong(eh.nextID))
	eh.nextID++
}

func (eh *JqErrorHandler) removeErrorHandler(jq *Jq) {
	delete(eh.callbacks, jq.errorHandlerID)
}

// This function is called for all errors in all jq instances.
// It therefore needs to dispatch the error to the correct jq instance
// error handler, based on the id parameter.
//export goJqErrorHandler
func goJqErrorHandler(id uint64, jv C.jv) {
	handler, ok := globalJqErrorHandler.callbacks[id]
	if ok {
		err := Jv{jv}
		handler(err.ToString())
	}
}

// New creates a new instance of Jq containing the jq state
func New() *Jq {
	jq := new(Jq)
	jq.state = C.jq_init()
	globalJqErrorHandler.addErrorHandler(jq)
	return jq
}

// Close frees the resources associated with this Jq instance
func (jq *Jq) Close() {
	C.jq_teardown(&jq.state)
	jq.state = nil
	globalJqErrorHandler.removeErrorHandler(jq)
}

func (jq *Jq) handleError(err string) {
	if jq.errorHandler != nil {
		jq.errorHandler(err)
	}
}

// CompileProgram compiles a jq program for the passed Jq instance
func (jq *Jq) CompileProgram(prog string) error {
	csProg := C.CString(prog)
	defer C.free(unsafe.Pointer(csProg))
	args := C.jv_array()

	errs := ""
	jq.errorHandler = func(err string) {
		errs += err + "\n"
	}

	C.jq_compile_args(jq.state, csProg, args)

	if errs != "" {
		return errors.New(errs)
	}
	return nil
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

func (jv *Jv) _toString(flags C.int) string {
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
	return jv._toString(C.int(0))
}

// PrettyPrint pretty prints the json value to stdout
func (jv *Jv) PrettyPrint() {
	prettyFlags := C.int(C.JV_PRINT_PRETTY | C.JV_PRINT_SPACE1)
	fmt.Printf("%s\n", jv._toString(prettyFlags))
}
