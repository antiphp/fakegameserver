package exiterror

import (
	"errors"
	"os"
	"strconv"
	"syscall"
	"time"
)

type ExitError struct {
	names   []string
	hookFns []func()
}

// New creates a new exit error with a signal and/or exit code.
func New(code, sig *int) *ExitError { // TODO: Weird signature; allow to configure separately and use syscall.Signal instead?
	if sig == nil && code == nil {
		return nil
	}

	exitErr := &ExitError{}
	if sig != nil {
		exitErr.addHook("signal "+strconv.Itoa(*sig), func() {
			_ = syscall.Kill(os.Getpid(), syscall.Signal(*sig))
			time.Sleep(5 * time.Second) // Brace for impact.
		})
	}
	if code != nil {
		exitErr.addHook("exit code "+strconv.Itoa(*code), func() {
			os.Exit(*code)
		})
	}
	return exitErr
}

func (e *ExitError) Error() string {
	for _, how := range e.names {
		return how
	}
	return "exit error"
}

func (e *ExitError) addHook(name string, hookFn func()) {
	e.hookFns = append(e.hookFns, hookFn)
	e.names = append(e.names, name)
}

func (e *ExitError) addHooks(names []string, hookFns []func()) {
	e.names = append(e.names, names...)
	e.hookFns = append(e.hookFns, hookFns...)
}

func (e *ExitError) RunHooks() {
	for _, hook := range e.hookFns {
		hook()
	}
}

// Wrap wraps multiple errors into a single exit error.
//
// Allows to execute runtime hooks first, then configured hooks, then default hooks.
func Wrap(errs ...error) *ExitError {
	retErr := &ExitError{}
	for _, err := range errs {
		if err == nil {
			continue
		}
		var exitErr *ExitError
		if errors.As(err, &exitErr) && exitErr != nil {
			retErr.addHooks(exitErr.names, exitErr.hookFns)
		}
	}
	return retErr
}
