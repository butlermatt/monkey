package compiler

type SymbolScope string

const (
	GlobalScope  SymbolScope = "GLOBAL"
	LocalScope   SymbolScope = "LOCAL"
	BuiltinScope SymbolScope = "BUILTIN"
	FreeScope    SymbolScope = "FREE"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	Outer *SymbolTable

	store       map[string]Symbol
	numDef      int
	FreeSymbols []Symbol
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{store: make(map[string]Symbol), FreeSymbols: []Symbol{}}
}

func NewEnclosedSymbolTable(table *SymbolTable) *SymbolTable {
	return &SymbolTable{Outer: table, store: make(map[string]Symbol)}
}

func (st *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{Name: name, Index: st.numDef}
	if st.Outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}

	st.store[name] = symbol
	st.numDef++
	return symbol
}

func (st *SymbolTable) Resolve(name string) (Symbol, bool) {
	s, ok := st.store[name]
	if ok || st.Outer == nil {
		return s, ok
	}

	s, ok = st.Outer.Resolve(name)
	if !ok || (s.Scope == GlobalScope || s.Scope == BuiltinScope) {
		return s, ok
	}

	free := st.defineFree(s)
	return free, true
}

func (st *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	sym := Symbol{Name: name, Index: index, Scope: BuiltinScope}
	st.store[name] = sym
	return sym
}

func (st *SymbolTable) defineFree(original Symbol) Symbol {
	st.FreeSymbols = append(st.FreeSymbols, original)

	sym := Symbol{Name: original.Name, Scope: FreeScope, Index: len(st.FreeSymbols) - 1}

	st.store[original.Name] = sym
	return sym
}
