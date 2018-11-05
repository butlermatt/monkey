package compiler

import "testing"

func TestDefine(t *testing.T) {
	expect := map[string]Symbol{
		"a": {Name: "a", Scope: GlobalScope, Index: 0},
		"b": {Name: "b", Scope: GlobalScope, Index: 1},
		"c": {Name: "c", Scope: LocalScope, Index: 0},
		"d": {Name: "d", Scope: LocalScope, Index: 1},
		"e": {Name: "e", Scope: LocalScope, Index: 0},
		"f": {Name: "f", Scope: LocalScope, Index: 1},
	}

	global := NewSymbolTable()

	a := global.Define("a")
	if a != expect["a"] {
		t.Errorf("expected a=%+v, got=%+v", expect["a"], a)
	}

	b := global.Define("b")
	if b != expect["b"] {
		t.Errorf("expected b=%+v, got=%+v", expect["b"], b)
	}

	firstLocal := NewEnclosedSymbolTable(global)

	c := firstLocal.Define("c")
	if c != expect["c"] {
		t.Errorf("expected c=%+v, got=%+v", expect["c"], c)
	}

	d := firstLocal.Define("d")
	if d != expect["d"] {
		t.Errorf("expected d=%+v, got=%+v", expect["d"], d)
	}

	secondLocal := NewEnclosedSymbolTable(firstLocal)

	e := secondLocal.Define("e")
	if e != expect["e"] {
		t.Errorf("expected e=%+v, got=%+v", expect["e"], e)
	}

	f := secondLocal.Define("f")
	if f != expect["f"] {
		t.Errorf("expected f=%+v, got=%+v", expect["f"], f)
	}
}

func TestResolveGlobal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	expect := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "b", Scope: GlobalScope, Index: 1},
	}

	for _, sym := range expect {
		t.Run(sym.Name, func(t *testing.T) {
			result, ok := global.Resolve(sym.Name)
			if !ok {
				t.Fatalf("unable to resolve name: %q", sym.Name)
			}
			if result != sym {
				t.Errorf("resolved symbol incorrect: expected=%+v, got=%+v", sym, result)
			}
		})
	}
}

func TestResolveLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	local := NewEnclosedSymbolTable(global)
	local.Define("c")
	local.Define("d")

	expect := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "b", Scope: GlobalScope, Index: 1},
		{Name: "c", Scope: LocalScope, Index: 0},
		{Name: "d", Scope: LocalScope, Index: 1},
	}

	for _, sym := range expect {
		t.Run(sym.Name, func(t *testing.T) {
			res, ok := local.Resolve(sym.Name)
			if !ok {
				t.Fatalf("unable to resolve name: %q", sym.Name)
			}

			if res != sym {
				t.Fatalf("resolved symbol incorrect: expected=%+v, got=%+v", sym, res)
			}
		})
	}
}

func TestResolveNestedLocals(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")
	firstLocal.Define("d")

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	tests := []struct {
		name     string
		table    *SymbolTable
		expected []Symbol
	}{
		{
			"first",
			firstLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
		},
		{
			"second",
			secondLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
				{Name: "e", Scope: LocalScope, Index: 0},
				{Name: "f", Scope: LocalScope, Index: 1},
			},
		},
	}

	for _, tt := range tests {
		for _, sym := range tt.expected {
			t.Run(tt.name+"-"+sym.Name, func(t *testing.T) {
				res, ok := tt.table.Resolve(sym.Name)
				if !ok {
					t.Fatalf("unable to resolve name: %q", sym.Name)
				}
				if res != sym {
					t.Fatalf("resolved symbol incorrect: expected=%+v, got=%+v", sym, res)
				}
			})
		}
	}
}

func TestDefineResolveBuiltins(t *testing.T) {
	global := NewSymbolTable()
	firstLocal := NewEnclosedSymbolTable(global)
	secondLocal := NewEnclosedSymbolTable(firstLocal)
	tables := []struct {
		name  string
		table *SymbolTable
	}{
		{"global", global},
		{"first", firstLocal},
		{"second", secondLocal},
	}

	expected := []Symbol{
		{Name: "a", Scope: BuiltinScope, Index: 0},
		{Name: "c", Scope: BuiltinScope, Index: 1},
		{Name: "e", Scope: BuiltinScope, Index: 2},
		{Name: "f", Scope: BuiltinScope, Index: 3},
	}

	for i, v := range expected {
		global.DefineBuiltin(i, v.Name)
	}

	for _, table := range tables {
		for _, sym := range expected {
			t.Run(table.name+"/"+sym.Name, func(t *testing.T) {
				res, ok := table.table.Resolve(sym.Name)
				if !ok {
					t.Fatalf("unable to resolve name: %q", sym.Name)
				}

				if res != sym {
					t.Errorf("resolved symbol %q incorrectly: expected=%+v, got=%+v", sym.Name, sym, res)
				}
			})
		}
	}
}
