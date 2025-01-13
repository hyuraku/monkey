package code

import "testing"

func TestMake(t *testing.T) {
	tests := []struct {
		opcode   Opcode
		operands []int
		expected []byte
	}{
		{OpConstant, []int{65534}, []byte{byte(OpConstant), 255, 254}},
	}

	for _, tt := range tests {
		instructions := Make(tt.opcode, tt.operands...)
		if len(instructions) != len(tt.expected) {
			t.Errorf("instructions has wrong length. got=%d, want=%d",
				len(instructions), len(tt.expected))
		}
		for i, b := range tt.expected {
			if instructions[i] != b {
				t.Errorf("at %d, wrong byte. got=%d, want=%d",
					i, instructions[i], b)
			}
		}
	}
}
