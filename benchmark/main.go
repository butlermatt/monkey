package main

import (
	"flag"
	"fmt"
	"github.com/butlermatt/monkey/compiler"
	"github.com/butlermatt/monkey/evaluator"
	"github.com/butlermatt/monkey/lexer"
	"github.com/butlermatt/monkey/object"
	"github.com/butlermatt/monkey/parser"
	"github.com/butlermatt/monkey/vm"
	"time"
)

var engine = flag.String("engine", "vm", "use 'vm' or 'eval'")
var input = `let fib = fn(x) {
	if (x == 0) {
		0
	} else {
		if (x == 1) {
			return 1;
		} else {
			fib(x - 1) + fib(x - 2);
		}
	}
};
fib(35);`

func main() {
	flag.Parse()

	var dur time.Duration
	var res object.Object

	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()

	if *engine == "vm" {
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			fmt.Printf("compiler error: %s\n", err)
			return
		}

		machine := vm.New(comp.ByteCode())

		start := time.Now()

		err = machine.Run()
		if err != nil {
			fmt.Printf("runtime error: %s\n", err)
			return
		}

		dur = time.Since(start)
		res = machine.LastPoppedStackElem()
	} else {
		env := object.NewEnvironment()
		start := time.Now()
		res = evaluator.Eval(program, env)
		dur = time.Since(start)
	}

	fmt.Printf("engine=%s, result=%s, duration=%s\n", *engine, res.Inspect(), dur)
}
