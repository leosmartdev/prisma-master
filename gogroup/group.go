// Package gogroup provides API to manage goroutines.
package gogroup

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"sync"
	"context"
)

type GoState int

const (
	GoStarted  GoState = iota
	GoFinished
)

var (
	// Set a callback function to be run whenever a goroutine starts or
	// completes
	Callback func(GoState)
)

/* There are two problems with goroutines which this package is intended to help solve:
 *
 *   1) When a goroutine panic()s and it is not caught, the entire application
 *   dies. This is a reasonable default behavior, but often one wants some
 *   other behavior.
 *
 *   2) One often launches a group of goroutines which cooperate. Should one of
 *   those goroutines die prematurely -- for whatever reason -- the others may
 *   simply stall infinitely, never making progress or releasing their
 *   resources.
 *
 * A GoGroup is a group of managed goroutines. When a new routine is launched,
 * it is protected from panic()s reaching the go runtime; instead, its panic()s
 * are caught, converted to errors and dealt with. Errors can simply be logged
 * with the routine restarted, or an error can lead to the entire group be
 * canceled.
 */
type GoGroup interface {
	context.Context

	// Cancel this group. Try to get all the children to exit
	Cancel(error)

	// Has this group been canceled?
	Canceled() bool

	// Launch a function in a new goroutine, but protected from panic()s
	// crashing everything. If it panic()s or returns an error, cancel this
	// Group, causing it to exit. If it exits normally, do nothing.
	Go(interface{}, ...interface{})

	// Launch a function in a new goroutine, but protected from panic()s
	// crashing everything. If it exits for any reason, restart it.
	GoRestart(interface{}, ...interface{})

	// Run a function in the current goroutine, but protected from panic()s
	// bubbling up beyond this point
	Run(interface{}, ...interface{})

	// Wait for all group threads to exit. Return all errors they threw
	Wait() []error

	// Iterate through the errors which have been thrown so far. Nil when no
	// more errors
	Error() error

	// Set a callback function to be run whenever an error is encountered
	ErrCallback(func(error))

	// Create a group which is a child context. Errors/panic()s in this child
	// do not affect the parent.
	Child(string) GoGroup

	Name() string
}

// An error converted from a recover()ed panic()
type PanicError struct {
	Msg   interface{}
	Stack string
}

func (pe PanicError) Error() string {
	if s, ok := pe.Msg.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", pe.Msg)
}

// Create a new group. Both arguments optional
func New(ctxt context.Context, name string) GoGroup {
	if ctxt == nil {
		ctxt = context.Background()
	}
	nctxt, cancel := context.WithCancel(ctxt)

	ret := group{
		Context: nctxt,
		name:    name,
		cancel:  cancel,
	}
	ret.ErrCallback(nil)
	return &ret
}

// Default implementation of a GoGroup
type group struct {
	context.Context
	sync.Mutex

	name   string
	cancel context.CancelFunc

	wg sync.WaitGroup

	errors      []error
	errCallback func(error)
	key, val    interface{}
}

// Cancel this group. Try to get all the children to exit
func (g *group) Cancel(err error) {
	if err != nil {
		g.errCallback(err)
	}
	g.cancel()
}

// Has this group been canceled?
func (g *group) Canceled() bool {
	select {
	case <-g.Done():
		return true
	default:
		return false
	}
}

// Launch a function in a new goroutine, but protected from panic()s
// crashing everything. If it panic()s or returns an error, cancel this
// Group, causing it to exit. If it exits normally, do nothing. Panic()s if
// argument is not runnable.
func (g *group) Go(f interface{}, args ...interface{}) {
	g.check(f, args...)
	g.wg.Add(1)
	go func() {
		g.run(f, args, true)
		g.wg.Done()
	}()
}

// Launch a function in a new goroutine, but protected from panic()s
// crashing everything. If it exits for any reason, restart it. Panic()s if
// argument is not runnable.
func (g *group) GoRestart(f interface{}, args ...interface{}) {
	g.check(f, args...)
	g.wg.Add(1)
	go func() {
		for !g.Canceled() {
			g.run(f, args, false)
		}
		g.wg.Done()
	}()
}

// Run a function in the current goroutine, but protected from panic()s
// bubbling up beyond this point. Panic()s if argument is not runnable.
func (g *group) Run(f interface{}, args ...interface{}) {
	g.check(f, args...)
	g.wg.Add(1)
	g.run(f, args, true)
	g.wg.Done()
}

func (g *group) check(f interface{}, args ...interface{}) {
	fv := reflect.ValueOf(f)
	ty := fv.Type()
	if ty.Kind() != reflect.Func {
		panic(fmt.Sprintf("Cannot run non-function value: %v", fv))
	}

	startsGroup := false
	if ty.NumIn() != 0 {
		inty := ty.In(0)
		if inty == reflect.TypeOf((*GoGroup)(nil)).Elem() {
			startsGroup = true
		}
	}

	tyInOffset := 0
	if len(args) != ty.NumIn() {
		if len(args)+1 == ty.NumIn() && startsGroup {
			tyInOffset = 1
		} else {
			panic(fmt.Sprintf("Function number of arguments doesn't match! (%v vs. %v)",
				len(args), ty.NumIn()))
		}
	}

	for i := 0; i < len(args); i++ {
		if !reflect.TypeOf(args[i]).AssignableTo(ty.In(i + tyInOffset)) {
			panic(fmt.Sprintf("Function argument number %v doesn't match! (%v vs. %v)",
				i, ty.In(i+tyInOffset), reflect.TypeOf(args[i])))
		}
	}
}

func (g *group) Name() string {
	return g.name
}

func (g *group) run(f interface{}, args []interface{}, kill bool) {
	defer g.catch(kill)

	if Callback != nil {
		Callback(GoStarted)
	}
	defer func() {
		if Callback != nil {
			Callback(GoFinished)
		}
	}()

	g.check(f, args...)
	fv := reflect.ValueOf(f)
	ty := fv.Type()
	if ty.Kind() != reflect.Func {
		panic(fmt.Sprintf("Cannot run non-function value: %v", fv))
	}

	// Run it!
	var vArgs []reflect.Value
	if len(args)+1 == ty.NumIn() {
		vArgs = append(vArgs, reflect.ValueOf(GoGroup(g)))
	}

	for _, arg := range args {
		vArgs = append(vArgs, reflect.ValueOf(arg))
	}

	var rets []reflect.Value
	if ty.IsVariadic() {
		rets = fv.CallSlice(vArgs)
	} else {
		rets = fv.Call(vArgs)
	}

	// Check results for an error
	for _, ret := range rets {
		if ret.IsValid() {
			err, ok := ret.Interface().(error)
			if ok && err != nil {
				g.errCallback(err)
			}
		}
	}
}

// Use this in a defer to catch panic()s
func (g *group) catch(kill bool) {
	if p := recover(); p != nil {
		// This function panic()ed!
		trace := debug.Stack()
		pe := PanicError{
			Msg:   p,
			Stack: string(trace),
		}

		if kill {
			g.Cancel(pe)
		} else {
			g.errCallback(pe)
		}
	}
}

// Wait for all group threads to exit. Return all errors they threw
func (g *group) Wait() []error {
	g.wg.Wait()
	g.Lock()
	defer g.Unlock()
	ret := g.errors
	g.errors = nil
	return ret
}

// Iterate through the errors which have been thrown so far. Nil when no
// more errors
func (g *group) Error() error {
	g.Lock()
	defer g.Unlock()
	if len(g.errors) == 0 {
		return nil
	}
	err := g.errors[0]
	g.errors = g.errors[1:]
	return err
}

// Child a child group
func (g *group) Child(name string) GoGroup {
	var ret *group
	if name == "" {
		ret = New(g, g.name+"-child").(*group)
	} else {
		ret = New(g, g.name+"-"+name).(*group)
	}
	ret.errCallback = g.errCallback
	return ret
}

// The default error handler
func (g *group) errorAppend(err error) {
	g.Lock()
	g.errors = append(g.errors, err)
	g.Unlock()
}

// Set a callback function to be run whenever an error is encountered
func (g *group) ErrCallback(f func(error)) {
	if f == nil {
		f = g.errorAppend
	}

	g.Lock()
	defer g.Unlock()
	g.errCallback = f
}

func WithValue(parent GoGroup, key, val interface{}) GoGroup {
	if key == nil {
		panic("nil key")
	}
	if !reflect.TypeOf(key).Comparable() {
		panic("key is not comparable")
	}
	original, ok := parent.(*group)
	if ok {
		return &group{
			Context:     parent,
			Mutex:       original.Mutex,
			name:        parent.Name(),
			cancel:      original.cancel,
			wg:          sync.WaitGroup{},
			errors:      original.errors,
			errCallback: original.errCallback,
			key:         key,
			val:         val,
		}
	}
	ctxt := context.WithValue(parent, key, val)
	return New(ctxt, parent.Name())
}

func (c *group) String() string {
	return fmt.Sprintf("%v.WithValue(%#v, %#v)", c.Context, c.key, c.val)
}

func (c *group) Value(key interface{}) interface{} {
	if c.key == key {
		return c.val
	}
	return c.Context.Value(key)
}
