package log

import (
	"context"
	"path"
	"runtime"
	"strconv"
	"strings"

	"github.com/cygy/ginamite/common/errors"
)

// logPackage is the import path of this package: its frames are never the caller of an error.
const logPackage = "github.com/cygy/ginamite/common/log"

// DatabaseError logs a database error with the collection it occurred on and its caller, see DatabaseOperationError.
// A nil or "not found" error is not logged. Prefer DatabaseOperationError, which also logs the operation and the
// attributes of the context.
func DatabaseError(err error, collection string) {
	if err == nil || errors.IsNotFound(err) {
		return
	}

	attrs := []any{"collection", collection}
	attrs = append(attrs, callerAttributes(false)...)

	Error("database error", append(attrs, "error", err.Error())...)
}

// DatabaseOperationError logs a database error with the operation which failed (such as "Aggregate"), the collection
// it occurred on, the attributes of the context (see ContextWithAttributes: the request ID, the route...) and its
// caller: the function which called the code reporting the error, so that the error leads to the code to fix. The
// caller is the first function outside this package and outside the package which calls DatabaseOperationError,
// reported as the attributes "caller" (the function) and "source" (the file and its line). A nil or "not found" error
// is not logged.
func DatabaseOperationError(ctx context.Context, err error, operation, collection string) {
	if err == nil || errors.IsNotFound(err) {
		return
	}

	attrs := []any{"operation", operation, "collection", collection}
	attrs = append(attrs, AttributesFromContext(ctx)...)

	attrs = append(attrs, callerAttributes(true)...)

	Error("database error", append(attrs, "error", err.Error())...)
}

// callerAttributes returns the attributes "caller" and "source" of the first frame outside this package, from the
// caller of the function which called callerAttributes. If skipReporter, the package of that caller, the code which
// reports the error, is skipped too: its wrappers are not the interesting caller.
func callerAttributes(skipReporter bool) []any {
	pcs := make([]uintptr, 16)
	// 3: runtime.Callers, callerAttributes and the function which logs.
	n := runtime.Callers(3, pcs)
	if n == 0 {
		return nil
	}

	frames := runtime.CallersFrames(pcs[:n])
	skipped := map[string]bool{logPackage: true}
	for first := true; ; first = false {
		frame, more := frames.Next()
		pkg := packageOf(frame.Function)
		if first && skipReporter {
			skipped[pkg] = true
		}

		if !skipped[pkg] {
			source := path.Base(path.Dir(frame.File)) + "/" + path.Base(frame.File) + ":" + strconv.Itoa(frame.Line)

			return []any{"caller", frame.Function, "source", source}
		}

		if !more {
			return nil
		}
	}
}

// packageOf returns the import path of the package of a function name such as
// "github.com/cygy/ginamite/common/log.DatabaseError" or "github.com/cygy/app/model.(*Score).Save".
func packageOf(function string) string {
	slash := strings.LastIndex(function, "/")
	dot := strings.Index(function[slash+1:], ".")
	if dot < 0 {
		return function
	}

	return function[:slash+1+dot]
}
