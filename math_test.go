package float128

import (
	"testing"
)

func TestMul(t *testing.T) {
	tests := []struct {
		a, b Float128
		want Float128
	}{
		{
			// 1 * 1 = 1
			Float128{0x3fff_0000_0000_0000, 0},
			Float128{0x3fff_0000_0000_0000, 0},
			Float128{0x3fff_0000_0000_0000, 0},
		},
		{
			// 1.5 * 1.5 = 2.25
			Float128{0x3fff_8000_0000_0000, 0},
			Float128{0x3fff_8000_0000_0000, 0},
			Float128{0x4000_2000_0000_0000, 0},
		},
		{
			// 2⁻¹⁶³⁸² * 2⁻¹ = 2⁻¹⁶³⁸³
			Float128{0x0001_0000_0000_0000, 0}, // smallest positive normal number
			Float128{0x3ffe_0000_0000_0000, 0}, // 0.5
			Float128{0x0000_8000_0000_0000, 0}, // 2⁻¹⁶³⁸³
		},
		{
			// underflow
			// 2⁻¹⁶⁴⁹⁴ * 2⁻¹ = 2⁻¹⁶⁴⁹⁵ ~ 0
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0001}, // smallest positive subnormal number
			Float128{0x3ffe_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.5
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000},
		},
		{
			// roundup subnormal
			// 2⁻¹⁶⁴⁹⁴ * 0.75 = 2⁻¹⁶⁴⁹⁵ ~ 0
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0001}, // smallest positive subnormal number
			Float128{0x3ffe_8000_0000_0000, 0x0000_0000_0000_0000}, // 0.75
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0001},
		},
		{
			// overflow
			Float128{0x7ffe_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff}, // largest normal number
			Float128{0x3fff_8000_0000_0000, 0x0000_0000_0000_0000}, // 1.5
			Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +Inf
		},

		// Infinity * 0 => NaN
		// 0 * Infinity => NaN
		{
			Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +Inf
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0
			Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0001}, // NaN
		},
		{
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0
			Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +Inf
			Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0001}, // NaN
		},

		// Infinity * anything => Infinity
		{
			// +Inf * +1 => +Inf
			Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +Inf
			Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +1
			Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +Inf
		},
		{
			// +Inf * -1 => -Inf
			Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +Inf
			Float128{0xBfff_0000_0000_0000, 0x0000_0000_0000_0000}, // -1
			Float128{0xffff_0000_0000_0000, 0x0000_0000_0000_0000}, // -Inf
		},
		{
			// -Inf * +1 => -Inf
			Float128{0xffff_0000_0000_0000, 0x0000_0000_0000_0000}, // -Inf
			Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +1
			Float128{0xffff_0000_0000_0000, 0x0000_0000_0000_0000}, // -Inf
		},
		{
			// -Inf * -1 => +Inf
			Float128{0xffff_0000_0000_0000, 0x0000_0000_0000_0000}, // -Inf
			Float128{0xBfff_0000_0000_0000, 0x0000_0000_0000_0000}, // -1
			Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +Inf
		},
	}

	for _, tt := range tests {
		got := tt.a.Mul(tt.b)
		if got != tt.want {
			t.Errorf("%#v * %#v: got %#v, want %#v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		a, b Float128
		want Float128
	}{
		{
			// 1 + 1 = 2
			Float128{0x3fff_0000_0000_0000, 0},
			Float128{0x3fff_0000_0000_0000, 0},
			Float128{0x4000_0000_0000_0000, 0},
		},
		{
			// 1 + 2 = 3
			Float128{0x3fff_0000_0000_0000, 0},
			Float128{0x4000_0000_0000_0000, 0},
			Float128{0x4000_8000_0000_0000, 0},
		},
	}

	for _, tt := range tests {
		got := tt.a.Add(tt.b)
		if got != tt.want {
			t.Errorf("%#v + %#v: got %#v, want %#v", tt.a, tt.b, got, tt.want)
		}
	}
}
