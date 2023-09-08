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
	}

	for _, tt := range tests {
		got := tt.a.Mul(tt.b)
		if got != tt.want {
			t.Errorf("%#v * %#v: got %#v, want %#v", tt.a, tt.b, got, tt.want)
		}
	}
}
