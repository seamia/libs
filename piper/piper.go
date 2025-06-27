package piper

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/seamia/libs/printer"
)

type (
	Processor func(string, printer.Printer)
)

func Run(proc Processor) {
	info, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if info.Mode()&os.ModeCharDevice != 0 || info.Size() <= 0 {
		fmt.Println("The command is intended to work with pipes.")
		os.Exit(123)
	}

	reader := bufio.NewReader(os.Stdin)
	var output []rune

	for {
		input, _, err := reader.ReadRune()
		if err != nil && err == io.EOF {
			break
		}
		output = append(output, input)
	}

	all := string(output)
	proc(all, printer.Stdout)
}
