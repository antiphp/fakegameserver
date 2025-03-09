package main

import (
	"errors"
	"os"
	"strconv"
	"syscall"
	"time"
)

type exitError struct {
	names   []string
	hookFns []func()
}

func (e *exitError) Error() string {
	for _, how := range e.names {
		return how
	}
	return "exit error"
}

func (e *exitError) addHook(name string, hookFn func()) {
	e.hookFns = append(e.hookFns, hookFn)
	e.names = append(e.names, name)
}

func (e *exitError) addHooks(names []string, hookFns []func()) {
	e.names = append(e.names, names...)
	e.hookFns = append(e.hookFns, hookFns...)
}

func (e *exitError) runHooks() {
	for _, hook := range e.hookFns {
		hook()
	}
}

// wrapExitErrs wraps multiple errors into a single exit error.
//
// Allows to execute runtime hooks first, then configured hooks, then default hooks.
func wrapExitErrs(errs ...error) *exitError {
	retErr := &exitError{}
	for _, err := range errs {
		if err == nil {
			continue
		}
		var exitErr *exitError
		if errors.As(err, &exitErr) && exitErr != nil {
			retErr.addHooks(exitErr.names, exitErr.hookFns)
		}
	}
	return retErr
}

// newExitErr creates a new exit error with a signal and/or exit code.
func newExitErr(code, sig *int) *exitError {
	if sig == nil && code == nil {
		return nil
	}

	exitErr := &exitError{}
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
