package printer

import (
	"fmt"
	"os"
)

type Printer func(format string, args ...interface{})

func Stdout(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stdout, format, args...)
}

func Stderr(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
}

type headerPrinter struct {
	header  string
	used    bool
	printer Printer
}

func (head *headerPrinter) print(format string, args ...interface{}) {
	if !head.used {
		head.used = true
		head.printer(head.header)
	}
	head.printer(format, args...)
}

// WithHeader - using this modifier you can create a 'printer' with attached 'header' that will be printed only if something else be printed
func (prn Printer) WithHeader(format string, args ...interface{}) Printer {
	printer := &headerPrinter{
		header:  fmt.Sprintf(format, args...),
		used:    false,
		printer: prn,
	}
	return printer.print
}

func empty(format string, args ...interface{}) {}

var (
	currentPrinter Printer = empty
)

func Print(format string, args ...interface{}) {
	currentPrinter(format, args...)
}

func Set(prn Printer) {
	if prn == nil {
		currentPrinter = empty
	} else {
		currentPrinter = prn
	}
}
