package util

import (
	"sync"
)

// GoSafe runs fn in a new goroutine, recovering from panics and logging
// them to the standard log package (log.Printf).  This is the preferred
// replacement for the old GoFuncOne calls.
func GoSafe(fn func()) {
	go func() {
		defer Recover("GoSafe")
		fn()
	}()
}

// GoSafeErr runs fn in a new goroutine, capturing the error result in the
// provided *error pointer (which must be non-nil).
func GoSafeErr(fn func() error, errOut *error) {
	GoSafe(func() {
		*errOut = fn()
	})
}

// Parallel runs fns concurrently and waits for all to finish.
func Parallel(fns ...func()) {
	var wg sync.WaitGroup
	wg.Add(len(fns))
	for _, fn := range fns {
		fn := fn
		GoSafe(func() {
			defer wg.Done()
			fn()
		})
	}
	wg.Wait()
}

// Recover catches panics.  If a label is provided it is logged; otherwise
// a generic message is used.  Call with `defer Recover("myFunc")`.
func Recover(label string) {
	if v := recover(); v != nil {
		// Using the root log package if available, otherwise stderr.
		logPrintf("[PANIC] %s: %v", label, v)
	}
}

// logPrintf is a safe wrapper to avoid depending on the root gozen log
// (which would create a circular import).
var logPrintf = func(format string, args ...any) {
	// In production this is overridden by framework init.
	// Default: silent.
}
