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

func TestInstructionsString(t *testing.T) {
	instructions := []Instructions{
		Make(OpConstant, 1),
		Make(OpConstant, 2),
		Make(OpConstant, 65534),
	}
	expected := `0000 OpConstant 1
0003 OpConstant 2
0006 OpConstant 65535
`
	concatted := Instructions{}
	for _, ins := range instructions {
		concatted = append(concatted, ins...)
	}

	if concatted.String() != expected {
		t.Errorf("instructions wrongly formatted.\nwant=%q\ngot=%q",
			expected, concatted.String())
	}

	}

