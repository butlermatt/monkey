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

func TestResolveFree(t *testing.T) {
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
		name    string
		table   *SymbolTable
		symbols []Symbol
		freeSym []Symbol
	}{
		{
			"first local",
			firstLocal,
			[]Symbol{
				{"a", GlobalScope, 0},
				{"b", GlobalScope, 1},
				{"c", LocalScope, 0},
				{"d", LocalScope, 1},
			},
			[]Symbol{},
		},
		{
			"second local",
			secondLocal,
			[]Symbol{
				{"a", GlobalScope, 0},
				{"b", GlobalScope, 1},
				{"c", FreeScope, 0},
				{"d", FreeScope, 1},
				{"e", LocalScope, 0},
				{"f", LocalScope, 1},
			},
			[]Symbol{
				{"c", LocalScope, 0},
				{"d", LocalScope, 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, sym := range tt.symbols {
				t.Run("expected symbols "+sym.Name, func(t *testing.T) {
					result, ok := tt.table.Resolve(sym.Name)
					if !ok {
						t.Fatalf("unable to resolve name: %q", sym.Name)
					}

					if result != sym {
						t.Fatalf("%q resolved wrong value. expected=%+v, got=%+v", sym.Name, sym, result)
					}
				})
			}

			if len(tt.table.FreeSymbols) != len(tt.freeSym) {
				t.Fatalf("wrong number of free symbols. expected=%d, got=%d", len(tt.freeSym), len(tt.table.FreeSymbols))
			}

			for i, sym := range tt.freeSym {
				t.Run("free symbols", func(t *testing.T) {
					res := tt.table.FreeSymbols[i]
					if res != sym {
						t.Fatalf("wrong symbol at index %d. expected=%+v, got=%+v", i, sym, res)
					}
				})
			}
		})
	}
}

func TestResolveUnresolvableFree(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	expected := []Symbol{
		{"a", GlobalScope, 0},
		{"c", FreeScope, 0},
		{"e", LocalScope, 0},
		{"f", LocalScope, 1},
	}

	for _, sym := range expected {
		t.Run("expected "+sym.Name, func(t *testing.T) {
			res, ok := secondLocal.Resolve(sym.Name)
			if !ok {
				t.Fatalf("unable to resolve name: %q", sym.Name)
			}

			if res != sym {
				t.Fatalf("symbol %q resolved to unexpected value. expected=%+v, got=%+v", sym.Name, sym, res)
			}
		})

		unexpected := []string{"b", "d"}
		for _, name := range unexpected {
			t.Run("unexpected "+name, func(t *testing.T) {
				_, ok := secondLocal.Resolve(name)
				if ok {
					t.Fatalf("resolved unexpected value: %q", name)
				}
			})
		}
	}
}
