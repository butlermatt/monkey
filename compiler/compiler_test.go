package compiler

import (
	"fmt"
	"github.com/butlermatt/monkey/ast"
	"github.com/butlermatt/monkey/code"
	"github.com/butlermatt/monkey/lexer"
	"github.com/butlermatt/monkey/object"
	"github.com/butlermatt/monkey/parser"
	"testing"
)

type compilerTestCase struct {
	name   string
	input  string
	consts []interface{}
	insts  []code.Instructions
}

func TestCompilerScopes(t *testing.T) {
	compiler := New()
	if compiler.scopeInd != 0 {
		t.Errorf("scope index wrong. expected=%d, got=%d", 0, compiler.scopeInd)
	}
	globalSymTable := compiler.symbolTable

	compiler.emit(code.OpMul)

	compiler.enterScope()
	if compiler.scopeInd != 1 {
		t.Errorf("scope index wrong. expected=%d, got=%d", 1, compiler.scopeInd)
	}

	compiler.emit(code.OpSub)

	if len(compiler.scopes[compiler.scopeInd].instructions) != 1 {
		t.Errorf("instructions length wrong. expected=%d, got=%d", 1, len(compiler.scopes[compiler.scopeInd].instructions))
	}

	last := compiler.scopes[compiler.scopeInd].last
	if last.Opcode != code.OpSub {
		t.Errorf("last instruction OpCode wrong. expected=%d, got=%d", code.OpSub, last.Opcode)
	}

	if compiler.symbolTable.Outer != globalSymTable {
		t.Errorf("compiler did not enclose symbol table")
	}

	compiler.leaveScope()
	if compiler.scopeInd != 0 {
		t.Errorf("scope index wrong. expected=%d, got=%d", 0, compiler.scopeInd)
	}

	if compiler.symbolTable != globalSymTable {
		t.Errorf("compiler did not restore global symbol table")
	}
	if compiler.symbolTable.Outer != nil {
		t.Errorf("compiler modified global symbol table incorrectly")
	}

	compiler.emit(code.OpAdd)

	if len(compiler.scopes[compiler.scopeInd].instructions) != 2 {
		t.Errorf("instruction length wrong. expected=%d, got=%d", 2, len(compiler.scopes[compiler.scopeInd].instructions))
	}

	last = compiler.scopes[compiler.scopeInd].last
	if last.Opcode != code.OpAdd {
		t.Errorf("last instruction OpCode wrong. expected=%d, got=%d", code.OpAdd, last.Opcode)
	}

	prev := compiler.scopes[compiler.scopeInd].prev
	if prev.Opcode != code.OpMul {
		t.Errorf("prev instruction OpCode wrong. expected=%d, got=%d", code.OpMul, last.Opcode)
	}
}

func TestNumberArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			name:   "one plus two",
			input:  "1 + 2;",
			consts: []interface{}{1.0, 2.0},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
		{
			name:   "one minus two",
			input:  "1 - 2;",
			consts: []interface{}{1.0, 2.0},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSub),
				code.Make(code.OpPop),
			},
		},
		{
			name:   "one times two",
			input:  "1 * 2;",
			consts: []interface{}{1.0, 2.0},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpMul),
				code.Make(code.OpPop),
			},
		},
		{
			name:   "two divided by 1",
			input:  "2 / 1;",
			consts: []interface{}{2.0, 1.0},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpDiv),
				code.Make(code.OpPop),
			},
		},
		{
			name:   "pop expression statement",
			input:  "1; 2",
			consts: []interface{}{1.0, 2.0},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			name:   "negative one",
			input:  "-1;",
			consts: []interface{}{1.0},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpMinus),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			name:   "true",
			input:  "true;",
			consts: []interface{}{},
			insts: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpPop),
			},
		},
		{
			name:   "false",
			input:  "false;",
			consts: []interface{}{},
			insts: []code.Instructions{
				code.Make(code.OpFalse),
				code.Make(code.OpPop),
			},
		},
		{
			name:   "1 Gt 2",
			input:  "1 > 2",
			consts: []interface{}{1.0, 2.0},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreater),
				code.Make(code.OpPop),
			},
		},
		{
			name:   "1 Lt 2",
			input:  "1 < 2",
			consts: []interface{}{2.0, 1.0},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreater),
				code.Make(code.OpPop),
			},
		},
		{
			name:   "1 GtEq 2",
			input:  "1 >= 2",
			consts: []interface{}{1.0, 2.0},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreaterEqual),
				code.Make(code.OpPop),
			},
		},
		{
			name:   "1 LtEq 2",
			input:  "1 <= 2",
			consts: []interface{}{2.0, 1.0},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreaterEqual),
				code.Make(code.OpPop),
			},
		},
		{
			name:   "1 EqEq 2",
			input:  "1 == 2",
			consts: []interface{}{1.0, 2.0},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpEqual),
				code.Make(code.OpPop),
			},
		},
		{
			name:   "1 NotEq 2",
			input:  "1 != 2",
			consts: []interface{}{1.0, 2.0},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpNotEqual),
				code.Make(code.OpPop),
			},
		},
		{
			name:   "true EqEq false",
			input:  "true == false",
			consts: []interface{}{},
			insts: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpFalse),
				code.Make(code.OpEqual),
				code.Make(code.OpPop),
			},
		},
		{
			name:   "true NotEq false",
			input:  "true != false",
			consts: []interface{}{},
			insts: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpFalse),
				code.Make(code.OpNotEqual),
				code.Make(code.OpPop),
			},
		},
		{
			name:   "not true",
			input:  "!true;",
			consts: []interface{}{},
			insts: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpBang),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []compilerTestCase{
		{
			name:   "if true ten",
			input:  `if (true) { 10 }; 3333;`,
			consts: []interface{}{10.0, 3333.0},
			insts: []code.Instructions{
				code.Make(code.OpTrue),            // 0000
				code.Make(code.OpJumpNotTrue, 10), // 0001
				code.Make(code.OpConstant, 0),     // 0004
				code.Make(code.OpJump, 11),        // 0007
				code.Make(code.OpNull),            // 0010
				code.Make(code.OpPop),             // 0011
				code.Make(code.OpConstant, 1),     // 0012
				code.Make(code.OpPop),             // 0015
			},
		},
		{
			name:   "if true ten else twenty",
			input:  `if (true) { 10 } else { 20 }; 3333;`,
			consts: []interface{}{10.0, 20.0, 3333.0},
			insts: []code.Instructions{
				code.Make(code.OpTrue),            // 0000
				code.Make(code.OpJumpNotTrue, 10), // 0001
				code.Make(code.OpConstant, 0),     // 0004
				code.Make(code.OpJump, 13),        // 0007
				code.Make(code.OpConstant, 1),     // 0010
				code.Make(code.OpPop),             // 0013
				code.Make(code.OpConstant, 2),     // 0014
				code.Make(code.OpPop),             // 0017
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestGlobalLetStatements(t *testing.T) {
	tests := []compilerTestCase{
		{
			name: "let one let two",
			input: `let one = 1;
let two = 2;
`,
			consts: []interface{}{1.0, 2.0},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSetGlobal, 1),
			},
		},
		{
			name: "let one and retrieve",
			input: `let one = 1;
one;`,
			consts: []interface{}{1.0},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			name: "let one let two is one",
			input: `let one = 1;
let two = one;
two;`,
			consts: []interface{}{1.0},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestLetStatementScopes(t *testing.T) {
	tests := []compilerTestCase{
		{
			name: "global in function",
			input: `let num = 55;
					fn() { num }`,
			consts: []interface{}{
				55.0,
				[]code.Instructions{
					code.Make(code.OpGetGlobal, 0),
					code.Make(code.OpReturnValue),
				},
			},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			name: "single local in function",
			input: `fn() {
						let num = 55;
						num
					}`,
			consts: []interface{}{
				55.0,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpReturnValue),
				},
			},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			name: "two locals in function return sum",
			input: `fn() {
						let a = 55;
						let b = 77;
						a + b
					}`,
			consts: []interface{}{
				55.0,
				77.0,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpSetLocal, 1),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			name:   "simple string",
			input:  `"monkey";`,
			consts: []interface{}{"monkey"},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			name:   "simple concatenation",
			input:  `"mon" + "key";`,
			consts: []interface{}{"mon", "key"},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			name:   "empty array",
			input:  "[]",
			consts: []interface{}{},
			insts: []code.Instructions{
				code.Make(code.OpArray, 0),
				code.Make(code.OpPop),
			},
		},
		{
			name:   "simple 3 numbers",
			input:  "[1, 2, 3]",
			consts: []interface{}{1.0, 2.0, 3.0},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpPop),
			},
		},
		{
			name:   "three simple expressions",
			input:  "[1 + 2, 3 - 4, 5 * 6]",
			consts: []interface{}{1.0, 2.0, 3.0, 4.0, 5.0, 6.0},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpSub),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpMul),
				code.Make(code.OpArray, 3),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestHashLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			name:   "empty hash",
			input:  "{}",
			consts: []interface{}{},
			insts: []code.Instructions{
				code.Make(code.OpHash, 0),
				code.Make(code.OpPop),
			},
		},
		{
			name:   "simple hash",
			input:  "{1: 2, 3: 4, 5: 6}",
			consts: []interface{}{1.0, 2.0, 3.0, 4.0, 5.0, 6.0},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpHash, 6),
				code.Make(code.OpPop),
			},
		},
		{
			name:   "complex hash values",
			input:  "{1: 2 + 3, 4: 5 * 6}",
			consts: []interface{}{1.0, 2.0, 3.0, 4.0, 5.0, 6.0},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpMul),
				code.Make(code.OpHash, 4),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			name:   "array one plus one index",
			input:  "[1, 2, 3][1 + 1]",
			consts: []interface{}{1.0, 2.0, 3.0, 1.0, 1.0},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpAdd),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
		{
			name:   "hash two minus one index",
			input:  "{1: 2}[2 - 1]",
			consts: []interface{}{1.0, 2.0, 2.0, 1.0},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpHash, 2),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpSub),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestFunctions(t *testing.T) {
	tests := []compilerTestCase{
		{
			name:  "no args explicit return",
			input: `fn() { return 5 + 10 }`,
			consts: []interface{}{
				5.0,
				10.0,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
		},
		{
			name:  "no args implicit return",
			input: `fn() { 5 + 10 }`,
			consts: []interface{}{
				5.0,
				10.0,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
		},
		{
			name:  "no args implicit return 2",
			input: `fn() { 1; 2 }`,
			consts: []interface{}{
				1.0,
				2.0,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpPop),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpReturnValue),
				},
			},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
		},
		{
			name:  "no args empty body",
			input: `fn() { }`,
			consts: []interface{}{
				[]code.Instructions{
					code.Make(code.OpReturn),
				},
			},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestFunctionCalls(t *testing.T) {
	tests := []compilerTestCase{
		{
			name:  "literal no args implicit return",
			input: `fn() { 24 }();`,
			consts: []interface{}{
				24.0,
				[]code.Instructions{
					code.Make(code.OpConstant, 0), // the literal 24
					code.Make(code.OpReturnValue),
				},
			},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 1), // compiled function
				code.Make(code.OpCall),
				code.Make(code.OpPop),
			},
		},
		{
			name: "call assigned function",
			input: `let noArg = fn() { 24 };
					noArg();`,
			consts: []interface{}{
				24.0,
				[]code.Instructions{
					code.Make(code.OpConstant, 0), // the literal 24
					code.Make(code.OpReturnValue),
				},
			},
			insts: []code.Instructions{
				code.Make(code.OpConstant, 1), // Compiled function.
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpCall),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func runCompilerTests(t *testing.T, tests []compilerTestCase) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := parse(tt.input)

			compiler := New()
			err := compiler.Compile(program)
			if err != nil {
				t.Fatalf("compiler error: %s", err)
			}

			bc := compiler.ByteCode()
			err = testInstructions(tt.insts, bc.Instructions)
			if err != nil {
				t.Fatalf("testInstructions failed: %s", err)
			}

			err = testConstants(t, tt.consts, bc.Constants)
			if err != nil {
				t.Fatalf("testConstants failed: %s", err)
			}
		})
	}
}

func testInstructions(expected []code.Instructions, actual code.Instructions) error {
	concatted := concatInstructions(expected)

	if len(actual) != len(concatted) {
		return fmt.Errorf("wrong instruction length.\nexpected=%q, got=%q", concatted, actual)
	}

	for i, ins := range concatted {
		if actual[i] != ins {
			return fmt.Errorf("wrong instruction at %d.\nexpected=%q, got=%q", i, ins, actual[i])
		}
	}

	return nil
}

func testConstants(t *testing.T, expected []interface{}, actual []object.Object) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("wrong number of constants. expected=%d, got=%d", len(expected), len(actual))
	}

	for i, constant := range expected {
		switch constant := constant.(type) {
		case float64:
			err := testNumberObject(constant, actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testNumberObject failed: %s", i, err)
			}
		case string:
			err := testStringObject(constant, actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testStringObject failed: %s", i, err)
			}
		case []code.Instructions:
			fn, ok := actual[i].(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("constant %d is not a function: %T", i, actual[i])
			}
			err := testInstructions(constant, fn.Instructions)
			if err != nil {
				return fmt.Errorf("constant %d - testInstructions failed: %s", i, err)
			}
		}
	}

	return nil
}

func testNumberObject(expected float64, actual object.Object) error {
	res, ok := actual.(*object.Number)
	if !ok {
		return fmt.Errorf("object is wrong type. expected=*object.Number got=%T (%+[1]v)", actual)
	}

	if res.Value != expected {
		return fmt.Errorf("object has wrong value. expected=%f, got=%f", expected, res.Value)
	}

	return nil
}

func testStringObject(expected string, actual object.Object) error {
	res, ok := actual.(*object.String)
	if !ok {
		return fmt.Errorf("object is wrong type. expected=*object.String, got=%T (%+[1]v)", actual)
	}

	if res.Value != expected {
		return fmt.Errorf("object has wrong value. expected=%q, got=%q", expected, res.Value)
	}

	return nil
}

func concatInstructions(s []code.Instructions) code.Instructions {
	var out code.Instructions

	for _, ins := range s {
		out = append(out, ins...)
	}

	return out
}
