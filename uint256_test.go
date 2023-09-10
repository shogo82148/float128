package float128

import (
	"testing"

	"github.com/shogo82148/int128"
)

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

func TestUint256DivMod(t *testing.T) {
	tests := []struct {
		a uint256
		b int128.Uint128
		q uint256
		r int128.Uint128
	}{
		{
			uint256{0x01234567_89abcdef, 0x01234567_89abcdef, 0x01234567_89abcdef, 0x01234567_89abcdef},
			int128.Uint128{L: 0x1_0000_0000},
			uint256{0x000000000_1234567, 0x89abcdef_01234567, 0x89abcdef_01234567, 0x89abcdef_01234567},
			int128.Uint128{L: 0x89abcdef},
		},
		{
			uint256{0x01234567_89abcdef, 0x01234567_89abcdef, 0x01234567_89abcdef, 0x01234567_89abcdef},
			int128.Uint128{H: 1},
			uint256{0, 0x01234567_89abcdef, 0x01234567_89abcdef, 0x01234567_89abcdef},
			int128.Uint128{L: 0x01234567_89abcdef},
		},
	}

	for _, tt := range tests {
		q, r := tt.a.divMod128(tt.b)
		if q != tt.q || r != tt.r {
			t.Errorf("%#v.divMod(%#v)\n got: %#v, %#v\nwant: %#v, %#v", tt.a, tt.b, q, r, tt.q, tt.r)
		}
	}
}
