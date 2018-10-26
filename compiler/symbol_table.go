package compiler

type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	store  map[string]Symbol
	numDef int
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{store: make(map[string]Symbol)}
}

func (st *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{Name: name, Index: st.numDef, Scope: GlobalScope}
	st.store[name] = symbol
	st.numDef++
	return symbol
}

func (st *SymbolTable) Resolve(name string) (Symbol, bool) {
	s, ok := st.store[name]
	return s, ok
}
