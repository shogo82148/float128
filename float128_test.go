package float128

import (
	"math"
	"testing"
)

// negZero is a float64 representation of -0.
var negZero = math.Float64frombits(1 << 63)

func TestIsNaN(t *testing.T) {
	tests := []struct {
		input Float128
		want  bool
	}{
		{NaN(), true},
		{Float128{0, 0}, false},
	}

	for _, tt := range tests {
		got := tt.input.IsNaN()
		if got != tt.want {
			t.Errorf("{%x, %x}.IsNaN() = %v, want %v", tt.input.h, tt.input.l, got, tt.want)
		}
	}
}

func TestIsInf(t *testing.T) {
	tests := []struct {
		input Float128
		sign  int
		want  bool
	}{
		{Inf(1), 1, true},
		{Inf(1), -1, false},
		{Inf(-1), 1, false},
		{Inf(-1), -1, true},

		{Inf(1), 0, true},
		{Inf(-1), 0, true},

		{Float128{0, 0}, 0, false},
	}

	for _, tt := range tests {
		got := tt.input.IsInf(tt.sign)
		if got != tt.want {
			t.Errorf("{%x, %x}.IsInf(%d) = %v, want %v", tt.input.h, tt.input.l, tt.sign, got, tt.want)
		}
	}
}

func TestFromFloat64(t *testing.T) {
	tests := []struct {
		input float64
		want  Float128
	}{
		{0, Float128{0, 0}},
		{negZero, Float128{0x8000_0000_0000_0000, 0}},
		{math.Inf(1), Float128{0x7fff_0000_0000_0000, 0}},
		{math.Inf(-1), Float128{0xffff_0000_0000_0000, 0}},
		{math.NaN(), Float128{0x7fff_8000_0000_0000, 0x01}},
		{1, Float128{0x3fff_0000_0000_0000, 0}},
		{-2, Float128{0xc000_0000_0000_0000, 0}},

		// small normal numbers of float64
		{0x1p-1022, Float128{0x3c01_0000_0000_0000, 0}},
		{0x1.0000000000001p-1022, Float128{0x3c01_0000_0000_0000, 0x1000_0000_0000_0000}},
		{0x1.fffffffffffffp-1022, Float128{0x3c01_ffff_ffff_ffff, 0xf000_0000_0000_0000}},

		// subnormal numbers of float64
		{0x1p-1023, Float128{0x3c00_0000_0000_0000, 0}},
		{0x1.0000000000002p-1023, Float128{0x3c00_0000_0000_0000, 0x2000_0000_0000_0000}},
		{0x1p-1074, Float128{0x3bcd_0000_0000_0000, 0}},
	}

	for _, tt := range tests {
		got := FromFloat64(tt.input)
		if got != tt.want {
			t.Errorf("FromFloat64(%x) = {0x%x, 0x%x}, want {0x%x, 0x%x}", tt.input, got.h, got.l, tt.want.h, tt.want.l)
		}
	}
}

func TestFloat64(t *testing.T) {
	tests := []struct {
		input Float128
		want  float64
	}{
		// special cases
		{Float128{0x7fff_0000_0000_0000, 0}, math.Inf(1)},
		{Float128{0xffff_0000_0000_0000, 0}, math.Inf(-1)},
		{Float128{0x7fff_8000_0000_0000, 0x01}, math.NaN()},

		// normal numbers
		{Float128{0x3fff_0000_0000_0000, 0}, 1},
		{Float128{0xc000_0000_0000_0000, 0}, -2},

		// small normal numbers of float64
		{Float128{0x3c01_0000_0000_0000, 0}, 0x1p-1022},
		{Float128{0x3c01_0000_0000_0000, 0x1000_0000_0000_0000}, 0x1.0000000000001p-1022},

		// subnormal numbers of float64
		{Float128{0x3c00_0000_0000_0000, 0}, 0x1p-1023},
		{Float128{0x3c00_0000_0000_0000, 0x1000_0000_0000_0000}, 0x1p-1023},
		{Float128{0x3c00_0000_0000_0000, 0x1000_0000_0000_0001}, 0x1.0000000000002p-1023},
		{Float128{0x3c00_0000_0000_0000, 0x3000_0000_0000_0000}, 0x1.0000000000004p-1023},
		{Float128{0x3bcd_0000_0000_0000, 0}, 0x1p-1074},

		// round to nearest, tie to even
		{Float128{0x3fff_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff}, 2},
	}

	for _, tt := range tests {
		got := tt.input.Float64()
		if math.Float64bits(got) != math.Float64bits(tt.want) {
			t.Errorf("{%x, %x}.Float64() = %x, want %x", tt.input.h, tt.input.l, got, tt.want)
		}
	}
}

func TestGoString(t *testing.T) {
	tests := []struct {
		input Float128
		want  string
	}{
		{Float128{0x7fff_0000_0000_0000, 0}, "+Inf"},
		{Float128{0xffff_0000_0000_0000, 0}, "-Inf"},
		{Float128{0x7fff_8000_0000_0000, 0x01}, "NaN"},
		{Float128{0x3fff_0000_0000_0000, 0}, "+0x1.0000000000000000000000000000p+0"},
		{Float128{0xc000_0000_0000_0000, 0}, "-0x1.0000000000000000000000000000p+1"},
		{Float128{0x3c01_0000_0000_0000, 0}, "+0x1.0000000000000000000000000000p-1022"},
		{Float128{0, 1}, "+0x0.0000000000000000000000000001p-16382"},
		{Float128{0, 0}, "+0x0.0000000000000000000000000000p+0"},
	}
	for _, tt := range tests {
		got := tt.input.GoString()
		if got != tt.want {
			t.Errorf("{%x, %x}.GoString() = %v, want %v", tt.input.h, tt.input.l, got, tt.want)
		}
	}
}
