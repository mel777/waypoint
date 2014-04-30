package main

import (
	"flag"
	"os"
	"fmt"
	"path/filepath"
	"runtime"
)

func Help() {
	Println("Usage of %s:", os.Args[0])
	flag.VisitAll(
		func(f *flag.Flag) {
			Println(" %-10s %-69s", f.Name, f.Usage)
		})
	os.Exit(0)
}

func ifError(err error) {
	if err != nil {
		Println(CaptureCaller(2) + " ERROR: %v", err)
        os.Exit(1)
	}
	return
}

func Fatal(msg string, a ...interface{}) {
	Println(CaptureCaller(2) + " FATAL: " + msg, a...)
    os.Exit(1)
	return
}

// Usually temporary for debugging
func Check(msg string, a ...interface{}) {
	Println("## CHECK " + msg, a...)
	return
}

func Println(msg string, a ...interface{}) {
	fmt.Printf(msg + "\n", a...)
	return
}

// Captures the callers details, accounting for jumps since the call.
func CaptureCaller(jumpsSinceCall int) string {
	_, filename, line, _ := runtime.Caller(jumpsSinceCall)
	return fmt.Sprintf("%s.%d", filepath.Base(filename), line)
}
