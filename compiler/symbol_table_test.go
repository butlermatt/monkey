package compiler

import "testing"

func TestDefine(t *testing.T) {
	expect := map[string]Symbol{
		"a": {Name: "a", Scope: GlobalScope, Index: 0},
		"b": {Name: "b", Scope: GlobalScope, Index: 1},
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
