package repl

import (
	"bufio"
	"fmt"
	"github.com/butlermatt/monkey/object"
	"io"

	"github.com/butlermatt/monkey/compiler"
	"github.com/butlermatt/monkey/lexer"
	"github.com/butlermatt/monkey/parser"
	"github.com/butlermatt/monkey/vm"
)

const Prompt = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	var consts []object.Object
	globals := make([]object.Object, vm.GlobalsSize)
	symbols := compiler.NewSymbolTable()

	for {
		fmt.Printf(Prompt)
		if scanned := scanner.Scan(); !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParseErrors(out, p.Errors())
			continue
		}

		comp := compiler.NewWithState(symbols, consts)
		err := comp.Compile(program)
		if err != nil {
			_, _ = fmt.Fprintf(out, "Woops! Compilation failed:\n%s\n", err)
			continue
		}

		code := comp.ByteCode()
		consts = code.Constants

		machine := vm.NewWithGlobalStore(code, globals)
		err = machine.Run()
		if err != nil {
			_, _ = fmt.Fprintf(out, "Woops! Executing bytecode failed:\n%s\n", err)
		}

		lastStack := machine.LastPoppedStackElem()
		_, _ = io.WriteString(out, lastStack.Inspect())
		_, _ = io.WriteString(out, "\n")
	}
}

func printParseErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		_, _ = io.WriteString(out, "\t"+msg+"\n")
	}
}
