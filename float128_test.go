package float128

import (
	"math"
	"math/bits"
	"runtime"
	"testing"
)

// negZero is a float64 representation of -0.
var negZero = math.Copysign(0, -1)

// xoshiro256++ PRNG
// It is used for benchmarking, avoiding branch prediction.
// Go port of http://prng.di.unimi.it/xoshiro256plusplus.c
type xoshiro256pp struct {
	s [4]uint64
}

func newXoshiro256pp() *xoshiro256pp {
	return &xoshiro256pp{
		[4]uint64{1, 2, 3, 4},
	}
}

func (s *xoshiro256pp) Uint64() uint64 {
	result := bits.RotateLeft64(s.s[0]+s.s[3], 23) + s.s[0]
	t := s.s[1] << 17
	s.s[2] ^= s.s[0]
	s.s[3] ^= s.s[1]
	s.s[1] ^= s.s[2]
	s.s[0] ^= s.s[3]
	s.s[2] ^= t
	s.s[3] = bits.RotateLeft64(s.s[3], 45)
	return result
}

func (s *xoshiro256pp) Float64() float64 {
	b := s.Uint64()
	return math.Float64frombits(b)
}

func (s *xoshiro256pp) Float128Pair() (Float128, Float128) {
	s.Uint64()
	return Float128{h: s.s[0], l: s.s[1]}, Float128{h: s.s[2], l: s.s[3]}
}

func BenchmarkXoshiro256ppFloat64(b *testing.B) {
	r := newXoshiro256pp()
	for i := 0; i < b.N; i++ {
		r.Float64()
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
		{math.NaN(), Float128{0x7fff_8000_0000_0000, 0x1000000000000000}},
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
		if !equals(got, tt.want) {
			t.Errorf("FromFloat64(%x) = {0x%x, 0x%x}, want {0x%x, 0x%x}", tt.input, got.h, got.l, tt.want.h, tt.want.l)
		}
	}
}

func BenchmarkFromFloat64(b *testing.B) {
	r := newXoshiro256pp()
	for i := 0; i < b.N; i++ {
		f := r.Float64()
		runtime.KeepAlive(FromFloat64(f))
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

		// overflow
		{Float128{0xfffe000000000000, 0x0}, math.Inf(-1)},
	}

	for _, tt := range tests {
		got := tt.input.Float64()
		if tt.input.IsNaN() && math.IsNaN(got) {
			continue
		}
		if math.Float64bits(got) != math.Float64bits(tt.want) {
			t.Errorf("{%x, %x}.Float64() = %x, want %x", tt.input.h, tt.input.l, got, tt.want)
		}
	}
}

func BenchmarkFloat64(b *testing.B) {
	r := newXoshiro256pp()
	for i := 0; i < b.N; i++ {
		f, _ := r.Float128Pair()
		runtime.KeepAlive(f.Float64())
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

func TestIsNaN(t *testing.T) {
	tests := []struct {
		input Float128
		want  bool
	}{
		// examples of NaN
		{NaN(), true},
		{Float128{0xffffffffffffffff, 0xfffffffffffffffe}, true},
		{Float128{0xffff000000000000, 0x0000000000000001}, true},

		{Float128{0, 0}, false},
	}

	for _, tt := range tests {
		got := tt.input.IsNaN()
		if got != tt.want {
			t.Errorf("{%x, %x}.IsNaN() = %v, want %v", tt.input.h, tt.input.l, got, tt.want)
		}
	}
}

// equals like a == b, but some exceptions.
//
//	NaN == NaN is true
//	-0 == 0 is false
func equals(a, b Float128) bool {
	if a.IsNaN() && b.IsNaN() {
		return true
	}
	return a == b
}
