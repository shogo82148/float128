package float128

import "testing"

func TestUint256Rsh(t *testing.T) {
	tests := []struct {
		input uint256
		n     uint
		want  uint256
	}{
		{
			uint256{0x01234567_89abcdef, 0x01234567_89abcdef, 0x01234567_89abcdef, 0x01234567_89abcdef},
			0,
			uint256{0x01234567_89abcdef, 0x01234567_89abcdef, 0x01234567_89abcdef, 0x01234567_89abcdef},
		},
		{
			uint256{0x01234567_89abcdef, 0x01234567_89abcdef, 0x01234567_89abcdef, 0x01234567_89abcdef},
			8,
			uint256{0x00012345_6789abcd, 0xef012345_6789abcd, 0xef012345_6789abcd, 0xef012345_6789abcd},
		},
		{
			uint256{0x01234567_89abcdef, 0x01234567_89abcdef, 0x01234567_89abcdef, 0x01234567_89abcdef},
			64,
			uint256{0x00000000_00000000, 0x01234567_89abcdef, 0x01234567_89abcdef, 0x01234567_89abcdef},
		},
	}

	for _, tt := range tests {
		got := tt.input.rsh(tt.n)
		if got != tt.want {
			t.Errorf("%#v.rsh(%d)\n got: %#v\nwant: %#v", tt.input, tt.n, got, tt.want)
		}
	}
}

func TestUint256Lsh(t *testing.T) {
	tests := []struct {
		input uint256
		n     uint
		want  uint256
	}{
		{
			uint256{0x01234567_89abcdef, 0x01234567_89abcdef, 0x01234567_89abcdef, 0x01234567_89abcdef},
			0,
			uint256{0x01234567_89abcdef, 0x01234567_89abcdef, 0x01234567_89abcdef, 0x01234567_89abcdef},
		},
		{
			uint256{0x01234567_89abcdef, 0x01234567_89abcdef, 0x01234567_89abcdef, 0x01234567_89abcdef},
			8,
			uint256{0x23456789_abcdef01, 0x23456789_abcdef01, 0x23456789_abcdef01, 0x23456789_abcdef00},
		},
		{
			uint256{0x01234567_89abcdef, 0x01234567_89abcdef, 0x01234567_89abcdef, 0x01234567_89abcdef},
			64,
			uint256{0x01234567_89abcdef, 0x01234567_89abcdef, 0x01234567_89abcdef, 0x00000000_00000000},
		},
	}

	for _, tt := range tests {
		got := tt.input.lsh(tt.n)
		if got != tt.want {
			t.Errorf("%#v.rsh(%d)\n got: %#v\nwant: %#v", tt.input, tt.n, got, tt.want)
		}
	}
}
