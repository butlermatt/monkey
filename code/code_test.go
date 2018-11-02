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
		{"OpAdd", OpAdd, []int{}, []byte{byte(OpAdd)}},
		{"OpGetLocal", OpGetLocal, []int{255}, []byte{byte(OpGetLocal), 255}},
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

func TestInstructionsString(t *testing.T) {
	inst := []Instructions{
		Make(OpAdd),
		Make(OpGetLocal, 1),
		Make(OpConstant, 2),
		Make(OpConstant, 65535),
	}

	expected := `0000 OpAdd
0001 OpGetLocal 1
0003 OpConstant 2
0006 OpConstant 65535
`

	concatted := Instructions{}
	for _, ins := range inst {
		concatted = append(concatted, ins...)
	}

	if concatted.String() != expected {
		t.Errorf("instructions formatted incorrectly. expected=%q, got=%q", expected, concatted.String())
	}
}

func TestReadOperands(t *testing.T) {
	tests := []struct {
		name     string
		op       OpCode
		operands []int
		read     int
	}{
		{"OpConstant", OpConstant, []int{65535}, 2},
		{"OpGetLocal", OpGetLocal, []int{255}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := Make(tt.op, tt.operands...)

			def, err := Lookup(byte(tt.op))
			if err != nil {
				t.Fatalf("Opcode definition not found: %q\n", err)
			}

			opsRead, n := ReadOperands(def, inst[1:])
			if n != tt.read {
				t.Fatalf("wrong number of bytes read. expected=%d, got=%d", tt.read, n)
			}

			for i, want := range tt.operands {
				if opsRead[i] != want {
					t.Errorf("wrong operand. expected=%d, got=%d", want, opsRead[i])
				}
			}
		})
	}
}
