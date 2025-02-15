package main

import "os"

type exitHookError struct {
	hookFn func()
}

func (e *exitHookError) appendExitCode(code int) *exitHookError {
	if e.hookFn != nil {
		return e
	}
	return &exitHookError{
		hookFn: func() {
			os.Exit(code)
		},
	}
}

func (e *exitHookError) Error() string {
	return "hook error"
}
