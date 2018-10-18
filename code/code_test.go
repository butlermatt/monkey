package code

import "testing"

func TestMake(t *testing.T) {
	tests := []struct {
		name     string
		op       OpCode
		operands []int
		expected []byte
	}{
		{"OpConstant", OpConstant, []int{65534}, []byte{byte(OpConstant), 255, 254}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := Make(tt.op, tt.operands...)

			if len(inst) != len(tt.expected) {
				t.Errorf("instruction has wrong length. expected=%d, got=%d", len(tt.expected), len(inst))
			}

			for i, b := range tt.expected {
				if inst[i] != b {
					t.Errorf("wrong byte at pos %d. expected=%d, got=%d", i, b, inst[i])
				}
			}
		})
	}
}
