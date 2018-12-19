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
import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"
)

// Jq represents a Jq state
type Jq struct {
	state          *C.struct_jq_state
	errorHandlerID uint64
	errorHandler   func(err string)
}

// Note: you can only set one error handler per jq instance

// ErrorCallbackMap keeps a map of error callbacks for jq instances.
type ErrorCallbackMap struct {
	nextID    uint64
	callbacks map[uint64]func(err string)
	lock      sync.RWMutex
}

var globalErrorCallbacks = ErrorCallbackMap{
	callbacks: make(map[uint64]func(err string)),
}

func (eh *ErrorCallbackMap) addErrorHandler(jq *Jq) {
	jq.errorHandlerID = atomic.AddUint64(&eh.nextID, 1)
	C.install_jq_error_cb(jq.state, C.ulonglong(jq.errorHandlerID))

	eh.lock.Lock()
	defer eh.lock.Unlock()
	eh.callbacks[jq.errorHandlerID] = jq.handleError
}

func (eh *ErrorCallbackMap) removeErrorHandler(jq *Jq) {
	eh.lock.Lock()
	defer eh.lock.Unlock()
	delete(eh.callbacks, jq.errorHandlerID)
}

func (eh *ErrorCallbackMap) getErrorHandler(id uint64) (func(err string), bool) {
	eh.lock.RLock()
	defer eh.lock.RUnlock()
	handler, ok := eh.callbacks[id]
	return handler, ok
}

// This function is called for all errors in all jq instances.
// It therefore needs to dispatch the error to the correct jq instance
// error handler, based on the id parameter.
//export goJqErrorHandler
func goJqErrorHandler(id uint64, jv C.jv) {
	handler, ok := globalErrorCallbacks.getErrorHandler(id)
	if ok {
		err := Jv{C.jq_format_error(jv)}
		handler(err.ToString())
	}
}

// New creates a new instance of Jq containing the jq state
func New() *Jq {
	jq := new(Jq)
	jq.state = C.jq_init()
	globalErrorCallbacks.addErrorHandler(jq)
	return jq
}

// Close frees the resources associated with this Jq instance
func (jq *Jq) Close() {
	C.jq_teardown(&jq.state)
	jq.state = nil
	globalErrorCallbacks.removeErrorHandler(jq)
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
func (jq *Jq) ProcessInput(input string) ([]*Jv, error) {
	jvInput, err := parseInput(input)
	if err != nil {
		return nil, err
	}
	defer C.jv_free(jvInput.jv)
	return jq.processInput(jvInput, false)
}

// ProcessInputFirstOnly runs the previously compiled program of the Jq instance on the input
// and only returns the first completed result.
func (jq *Jq) ProcessInputFirstOnly(input string) ([]*Jv, error) {
	jvInput, err := parseInput(input)
	if err != nil {
		return nil, err
	}
	defer C.jv_free(jvInput.jv)
	return jq.processInput(jvInput, true)
}

func (jq *Jq) ProcessInputJv(input *Jv) ([]*Jv, error) {
	return jq.processInput(input.Copy(), false)
}

func (jq *Jq) processInput(jvInput *Jv, firstOnly bool) ([]*Jv, error) {
	results := make([]*Jv, 0)
	flags := C.int(0)

	C.jq_start(jq.state, jvInput.jv, flags)
	res := C.jq_next(jq.state)
	for C.jv_is_valid(res) != 0 {
		results = append(results, &Jv{res})
		if firstOnly {
			break
		}
		res = C.jq_next(jq.state)
	}

	return results, nil
}

func parseInput(input string) (*Jv, error) {
	csInput := C.CString(input)
	defer C.free(unsafe.Pointer(csInput))

	jvInput := C.jv_parse(csInput)
	if C.jv_is_valid(jvInput) == 0 {
		err := Jv{C.jq_format_error(jvInput)}
		return nil, errors.New(err.ToString())
	}
	return &Jv{jvInput}, nil
}

// Jv represents a Jq json value (jv)
type Jv struct {
	jv C.jv
}

func (jv *Jv) Copy() *Jv {
	C.jv_copy(jv.jv)
	return jv
}

func (jv *Jv) Free() *Jv {
	// Cf. https://github.com/stedolan/jq/wiki/C-API:-jv#memory-management
	C.jv_free(jv.jv)
	return nil
}

func (jv *Jv) IsNumber() bool {
	return C.jv_get_kind(jv.jv) == C.JV_KIND_NUMBER
}

func (jv *Jv) IsString() bool {
	return C.jv_get_kind(jv.jv) == C.JV_KIND_STRING
}

func (jv *Jv) ToNumber() interface{} {
	dbl := C.jv_number_value(jv.jv)
	if C.jv_is_integer(jv.jv) == 0 {
		return float64(dbl)
	}
	return int(dbl)
}

func (jv *Jv) _toString(flags C.int) string {
	jvStr := C.jv_dump_string(jv.Copy().jv, flags)
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
