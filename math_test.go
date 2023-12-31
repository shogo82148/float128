package float128

import (
	"fmt"
	"runtime"
	"testing"
)

func dump(f Float128) string {
	return fmt.Sprintf("%#v (%016x_%016x)", f, f.h, f.l)
}

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
			Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
		},
		{
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0
			Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +Inf
			Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
		},

		// 0 * anything => 0
		{
			Float128{0x4280_0000_01ff_ffff, 0xffff_f7ff_ffff_ffff}, // +0x1.000001fffffffffff7ffffffffffp+641
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0
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

		// edge case of normal and subnormal number
		{
			Float128{0x0000ffffffffffff, 0xffffffffffffffff}, // +0x0.ffffffffffffffffffffffffffffp-16382
			Float128{0x3fff000000000000, 0x0000000000000001}, // +0x1.0000000000000000000000000001p+0
			Float128{0x0001000000000000, 0x0000000000000000}, // +0x1.0000000000000000000000000000p-16382
		},
	}

	for _, tt := range tests {
		got := tt.a.Mul(tt.b)
		if !equals(got, tt.want) {
			t.Errorf("%s * %s: got %s, want %s", dump(tt.a), dump(tt.b), dump(got), dump(tt.want))
		}
	}
}

func BenchmarkMul(b *testing.B) {
	r := newXoshiro256pp()
	for i := 0; i < b.N; i++ {
		a, b := r.Float128Pair()
		runtime.KeepAlive(a.Mul(b))
	}
}

func TestQuo(t *testing.T) {
	tests := []struct {
		a, b Float128
		want Float128
	}{
		// normal number / normal number
		{
			// 1 / 1 = 1
			Float128{0x3fff_0000_0000_0000, 0},
			Float128{0x3fff_0000_0000_0000, 0},
			Float128{0x3fff_0000_0000_0000, 0},
		},
		{
			// 1 / 2 = 0.5
			Float128{0x3fff_0000_0000_0000, 0},
			Float128{0x4000_0000_0000_0000, 0},
			Float128{0x3ffe_0000_0000_0000, 0},
		},
		{
			// 1 / 3
			Float128{0x3fff_0000_0000_0000, 0},
			Float128{0x4000_8000_0000_0000, 0},
			Float128{0x3ffd_5555_5555_5555, 0x5555_5555_5555_5555},
		},

		// the result is subnormal
		{
			// 2⁻¹⁶³⁸² / 2 = 2⁻¹⁶³⁸³
			Float128{0x0001_0000_0000_0000, 0}, // smallest positive normal number
			Float128{0x4000_0000_0000_0000, 0}, // 2
			Float128{0x0000_8000_0000_0000, 0}, // 2⁻¹⁶³⁸³
		},

		// underflow
		{
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0001}, // smallest positive subnormal number
			Float128{0x4000_0000_0000_0000, 0},                     // 2
			Float128{0, 0},                                         // 0
		},

		// overflow
		{
			Float128{0x7ffe_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff}, // largest normal number
			Float128{0x3ffe_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.5
			Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +Inf
		},

		// anything / 0
		{
			// +1 / +0 => +Inf
			Float128{0x3fff_0000_0000_0000, 0}, // 1
			Float128{0x0000_0000_0000_0000, 0}, // +0
			Float128{0x7fff_0000_0000_0000, 0}, // +Inf
		},
		{
			// +1 / -0 => +Inf
			Float128{0x3fff_0000_0000_0000, 0}, // 1
			Float128{0x8000_0000_0000_0000, 0}, // -0
			Float128{0xffff_0000_0000_0000, 0}, // -Inf
		},
		{
			// -1 / +0 => +Inf
			Float128{0xbfff_0000_0000_0000, 0}, // -1
			Float128{0x0000_0000_0000_0000, 0}, // +0
			Float128{0xffff_0000_0000_0000, 0}, // -Inf
		},
		{
			// -1 / -0 => +Inf
			Float128{0xbfff_0000_0000_0000, 0}, // -1
			Float128{0x8000_0000_0000_0000, 0}, // -0
			Float128{0x7fff_0000_0000_0000, 0}, // +Inf
		},

		// 0 / anything => 0
		{
			Float128{0x0000_0000_0000_0000, 0}, // +0
			Float128{0x3fff_0000_0000_0000, 0}, // 1
			Float128{0x0000_0000_0000_0000, 0}, // +0
		},
		{
			Float128{0x8000_0000_0000_0000, 0}, // -0
			Float128{0x3fff_0000_0000_0000, 0}, // 1
			Float128{0x8000_0000_0000_0000, 0}, // -0
		},

		// anything / Inf
		{
			Float128{0x3fff_0000_0000_0000, 0}, // 1
			Float128{0x7fff_0000_0000_0000, 0}, // +Inf
			Float128{0x0000_0000_0000_0000, 0}, // +0
		},

		// NaN / anything => NaN
		{
			Float128{0x7fff_8000_0000_0000, 0x01}, // NaN
			Float128{0x7fff_8000_0000_0000, 0x01}, // NaN
			Float128{0x7fff_8000_0000_0000, 0x01}, // NaN
		},
		{
			Float128{0x7fff_8000_0000_0000, 0x01}, // NaN
			Float128{0x3fff_0000_0000_0000, 0},    // 1
			Float128{0x7fff_8000_0000_0000, 0x01}, // NaN
		},

		// anything / NaN => NaN
		{
			Float128{0x3fff_0000_0000_0000, 0},    // 1
			Float128{0x7fff_8000_0000_0000, 0x01}, // NaN
			Float128{0x7fff_8000_0000_0000, 0x01}, // NaN
		},

		{
			Float128{0x4001e94284905f1d, 0xa44c3a217be5490c}, // +0x1.e94284905f1da44c3a217be5490cp+2
			Float128{0xc08000003fffffff, 0xfffffeffffffffff}, // 0x1.00003ffffffffffffeffffffffffp+129
			Float128{0xbf80e9420a3fdc8d, 0xad28d0c089bf667b}, // -0x1.e9420a3fdc8dad28d0c089bf667bp-127
		},
		{
			Float128{0x0001007fffffffff, 0xffefffffffffffff}, // +0x1.007fffffffffffefffffffffffffp-16382
			Float128{0x4000fffffffc0000, 0x0000000000040000}, // +0x1.fffffffc00000000000000040000p+1
			Float128{0x000040200000803f, 0xfffd007ffff980c0}, // +0x0.40200000803ffffd007ffff980c0p-16382
		},
	}

	for _, tt := range tests {
		got := tt.a.Quo(tt.b)
		if !equals(got, tt.want) {
			t.Errorf("%s / %s: got %s, want %s", dump(tt.a), dump(tt.b), dump(got), dump(tt.want))
		}
	}
}

func BenchmarkQuo(b *testing.B) {
	r := newXoshiro256pp()
	for i := 0; i < b.N; i++ {
		a, b := r.Float128Pair()
		runtime.KeepAlive(a.Quo(b))
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		a, b Float128
		want Float128
	}{
		// normal number + normal number
		{
			// 1 + 1 = 2
			Float128{0x3fff_0000_0000_0000, 0},
			Float128{0x3fff_0000_0000_0000, 0},
			Float128{0x4000_0000_0000_0000, 0},
		},
		{
			// (-1) + (-1) = (-2)
			Float128{0xbfff_0000_0000_0000, 0},
			Float128{0xbfff_0000_0000_0000, 0},
			Float128{0xc000_0000_0000_0000, 0},
		},
		{
			// 1 + 2 = 3
			Float128{0x3fff_0000_0000_0000, 0},
			Float128{0x4000_0000_0000_0000, 0},
			Float128{0x4000_8000_0000_0000, 0},
		},
		{
			// 2 + 1 = 3
			Float128{0x4000_0000_0000_0000, 0},
			Float128{0x3fff_0000_0000_0000, 0},
			Float128{0x4000_8000_0000_0000, 0},
		},
		{
			// 1 + 2⁻¹¹²
			Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000},
			Float128{0x3f8f_0000_0000_0000, 0x0000_0000_0000_0000},
			Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0001}, // smallest number larger than one
		},
		{
			// round up
			// 1 + 1.75 * 2⁻¹¹² ~ 1 + 2 * 2⁻¹¹²
			Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000},
			Float128{0x3f8f_c000_0000_0000, 0x0000_0000_0000_0000},
			Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0002},
		},
		{
			// round down
			// 1 + 2⁻¹¹³ ~ 1
			Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000},
			Float128{0x3f8e_0000_0000_0000, 0x0000_0000_0000_0000},
			Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000},
		},
		{
			// round down
			// 1 + 1.0000000000000000000000000000000001926 * 2⁻¹¹³ ~ 1.0000000000000000000000000000000001926
			Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000},
			Float128{0x3f8e_0000_0000_0000, 0x0000_0000_0000_0001},
			Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0001},
		},
		{
			// overflow
			Float128{0x7ffe_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff}, // 2¹⁶³⁸³ × (2 − 2⁻¹¹²)
			Float128{0x7f8e_0000_0000_0000, 0x0000_0000_0000_0000}, // 2¹⁶³⁸³ × 2⁻¹¹²
			Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +Inf
		},

		// subnormal number + subnormal number => normal number
		{
			Float128{0x0000_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff}, // largest subnormal number
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0001}, // smallest positive subnormal number
			Float128{0x0001_0000_0000_0000, 0x0000_0000_0000_0000},
		},
		{
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0001}, // smallest positive subnormal number
			Float128{0x0000_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff}, // largest subnormal number
			Float128{0x0001_0000_0000_0000, 0x0000_0000_0000_0000},
		},

		// subnormal number + subnormal number => subnormal number
		{
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0001}, // smallest positive subnormal number
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0001}, // smallest positive subnormal number
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0002},
		},

		// positive normal number + negative normal number => normal number
		{
			// 1 + (-0.5) = 0.5
			Float128{0x3fff_0000_0000_0000, 0},
			Float128{0xbffe_0000_0000_0000, 0},
			Float128{0x3ffe_0000_0000_0000, 0},
		},
		{
			// (-0.5) + 1 = 0.5
			Float128{0xbffe_0000_0000_0000, 0},
			Float128{0x3fff_0000_0000_0000, 0},
			Float128{0x3ffe_0000_0000_0000, 0},
		},
		{
			// 1 + (-2⁻¹¹³)
			Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000},
			Float128{0xbf8e_0000_0000_0000, 0x0000_0000_0000_0000},
			Float128{0x3ffe_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff}, // largest number less than one
		},
		{
			// (-2⁻¹¹³) + 1
			Float128{0xbf8e_0000_0000_0000, 0x0000_0000_0000_0000},
			Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000},
			Float128{0x3ffe_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff}, // largest number less than one
		},

		// 0 + 0 => 0
		{
			// (+0) + (+0) = +0
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000},
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000},
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000},
		},
		{
			// (+0) + (-0) = +0
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000},
			Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000},
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000},
		},
		{
			// (-0) + (+0) = +0
			Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000},
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000},
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000},
		},
		{
			// (-0) + (-0) = -0
			Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000},
			Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000},
			Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000},
		},

		// 0 + anything => anything
		{
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000},
			Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1
			Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1
		},

		// random tests
		{
			Float128{0x3c01001000000000, 0x0000000ffffffffe}, // +0x1.0010000000000000000ffffffffep-1022
			Float128{0xbc01000000000000, 0x00002000003fffff}, // +0x1.0010000000000000000ffffffffep-1022
			Float128{0x3bf4ffffffffffff, 0xfc01fff7ffffe000}, // +0x1.fffffffffffffc01fff7ffffe000p-1035
		},
		{
			Float128{0xc00071d51111ce6a, 0x4509d1b333c83266}, // -0x1.71d51111ce6a4509d1b333c83266p+1
			Float128{0x3ffd04b061817585, 0x2e6a396ed009b694}, // +0x1.04b0618175852e6a396ed009b694p-2
			Float128{0xc000513f04e19fb9, 0x9f3c8a8559c6fb94}, // -0x1.513f04e19fb99f3c8a8559c6fb94p+1
		},
		{
			Float128{0x4002ffffffc00000, 0x0010000000000000}, // +0x1.ffffffc000000010000000000000p+3
			Float128{0xc000ffffffffffff, 0xfffffffffffffffe}, // -0x1.fffffffffffffffffffffffffffep+1
			Float128{0x40027fffffc00000, 0x0010000000000000}, // +0x1.7fffffc000000010000000000000p+3
		},
		{
			Float128{0x0001ffffe0000000, 0x003ffffffffffffe}, // +0x1.ffffe0000000003ffffffffffffep-16382
			Float128{0x8001ffffffffffff, 0xfffffffffffffffe}, // -0x1.fffffffffffffffffffffffffffep-16382
			Float128{0x800000001fffffff, 0xffc0000000000000}, // -0x0.00001fffffffffc0000000000000p-16382
		},
		{
			Float128{0x3f8e000000000000, 0x0000000000000000}, // +0x1.0000000000000000000000000000p-113
			Float128{0xbf8e000000000000, 0x0000000000000000}, // -0x1.0000000000000000000000000000p-113
			Float128{0x0000000000000000, 0x0000000000000000}, // 0
		},
	}

	for _, tt := range tests {
		got := tt.a.Add(tt.b)
		if !equals(got, tt.want) {
			t.Errorf("%s + %s: got %s, want %s", dump(tt.a), dump(tt.b), dump(got), dump(tt.want))
		}
	}
}

func BenchmarkAdd(b *testing.B) {
	r := newXoshiro256pp()
	for i := 0; i < b.N; i++ {
		a, b := r.Float128Pair()
		runtime.KeepAlive(a.Add(b))
	}
}

func TestComparison(t *testing.T) {
	tests := []struct {
		a, b                   Float128
		eq, ne, lt, gt, le, ge bool
	}{
		{
			Float128{0x3fff_0000_0000_0000, 0}, // 1
			Float128{0x3fff_0000_0000_0000, 0}, // 1
			true, false, false, false, true, true,
		},
		{
			Float128{0x3fff_0000_0000_0000, 0}, // 1
			Float128{0x0000_0000_0000_0000, 0}, // 0
			false, true, false, true, false, true,
		},
		{
			Float128{0x0000_0000_0000_0000, 0}, // 0
			Float128{0x3fff_0000_0000_0000, 0}, // 1
			false, true, true, false, true, false,
		},
		{
			Float128{0x0000_0000_0000_0000, 0}, // +0
			Float128{0x8000_0000_0000_0000, 0}, // -0
			true, false, false, false, true, true,
		},

		// NaN
		{
			Float128{0x7fff_8000_0000_0000, 0x01}, // NaN
			Float128{0x3fff_0000_0000_0000, 0},    // 1
			false, true, false, false, false, false,
		},
	}

	for _, tt := range tests {
		eq := tt.a.Eq(tt.b)
		if eq != tt.eq {
			t.Errorf("%s == %s: got %t, want %t", dump(tt.a), dump(tt.b), eq, tt.eq)
		}

		ne := tt.a.Ne(tt.b)
		if ne != tt.ne {
			t.Errorf("%s != %s: got %t, want %t", dump(tt.a), dump(tt.b), ne, tt.ne)
		}

		lt := tt.a.Lt(tt.b)
		if lt != tt.lt {
			t.Errorf("%s < %s: got %t, want %t", dump(tt.a), dump(tt.b), lt, tt.lt)
		}

		gt := tt.a.Gt(tt.b)
		if gt != tt.gt {
			t.Errorf("%s > %s: got %t, want %t", dump(tt.a), dump(tt.b), gt, tt.gt)
		}

		le := tt.a.Le(tt.b)
		if le != tt.le {
			t.Errorf("%s <= %s: got %t, want %t", dump(tt.a), dump(tt.b), le, tt.le)
		}

		ge := tt.a.Ge(tt.b)
		if ge != tt.ge {
			t.Errorf("%s >= %s: got %t, want %t", dump(tt.a), dump(tt.b), ge, tt.ge)
		}
	}
}

func BenchmarkEq(b *testing.B) {
	r := newXoshiro256pp()
	for i := 0; i < b.N; i++ {
		a, b := r.Float128Pair()
		runtime.KeepAlive(a.Eq(b))
	}
}

func BenchmarkLt(b *testing.B) {
	r := newXoshiro256pp()
	for i := 0; i < b.N; i++ {
		a, b := r.Float128Pair()
		runtime.KeepAlive(a.Lt(b))
	}
}

func BenchmarkLe(b *testing.B) {
	r := newXoshiro256pp()
	for i := 0; i < b.N; i++ {
		a, b := r.Float128Pair()
		runtime.KeepAlive(a.Le(b))
	}
}

func TestFMA(t *testing.T) {
	tests := []struct {
		x, y, z Float128
		want    Float128
	}{
		// NaN
		{
			Float128{0x7fff_8000_0000_0000, 0x01}, // NaN
			Float128{0x7fff_8000_0000_0000, 0x01}, // NaN
			Float128{0x7fff_8000_0000_0000, 0x01}, // NaN
			Float128{0x7fff_8000_0000_0000, 0x01}, // NaN
		},

		// 1 * 1 + 1 = 2
		{
			Float128{0x3fff_0000_0000_0000, 0}, // 1
			Float128{0x3fff_0000_0000_0000, 0}, // 1
			Float128{0x3fff_0000_0000_0000, 0}, // 1
			Float128{0x4000_0000_0000_0000, 0}, // 2
		},

		// 1 * 1 + 0.5 = 1.5
		{
			Float128{0x3fff_0000_0000_0000, 0}, // 1
			Float128{0x3fff_0000_0000_0000, 0}, // 1
			Float128{0x3ffe_0000_0000_0000, 0}, // 0.5
			Float128{0x3fff_8000_0000_0000, 0}, // 1.5
		},

		// 1 * 1 + (-0.5) = 0.5
		{
			Float128{0x3fff_0000_0000_0000, 0}, // 1
			Float128{0x3fff_0000_0000_0000, 0}, // 1
			Float128{0xbffe_0000_0000_0000, 0}, // 0.5
			Float128{0x3ffe_0000_0000_0000, 0}, // 0.5
		},

		// the result is subnormal
		{
			Float128{0x0001_0000_0000_0000, 0}, // smallest positive normal number
			Float128{0x3ffe_0000_0000_0000, 0}, // 0.5
			Float128{0x0000_0000_0000_0000, 1}, // // smallest positive subnormal number
			Float128{0x0000_8000_0000_0000, 1}, // 2⁻¹⁶³⁸³ + 2⁻¹⁶⁴⁹⁴
		},
		{
			Float128{0x0001_0000_0000_0000, 0}, // smallest positive normal number
			Float128{0x3ffd_0000_0000_0000, 0}, // 0.25
			Float128{0x0000_0000_0000_0000, 1}, // // smallest positive subnormal number
			Float128{0x0000_4000_0000_0000, 1}, // 2⁻¹⁶³⁸³ + 2⁻¹⁶⁴⁹⁴
		},
		{
			Float128{0xbffdfffffffff7ff, 0xffffffffefffffff}, // -0x1.fffffffff7ffffffffffefffffffp-2
			Float128{0x0000fffffffffffb, 0xffffffffffffffef}, // +0x0.fffffffffffbffffffffffffffefp-16382
			Float128{0x000000fde9f01372, 0x332c149127a0cf42}, // +0x0.00fde9f01372332c149127a0cf42p-16382
			Float128{0x80007f02160fea8b, 0xccd3eb6edc5f30b5}, // -0x0.7f02160fea8bccd3eb6edc5f30b5p-16382
		},
		{
			Float128{0x8001fffff807ffff, 0xfffffffffffffffe},
			Float128{0xbf7eefc46289dd61, 0x3b5141a1d876a1ea},
			Float128{0x0000000fffffffff, 0xfffffffffff00000},
			Float128{0x0000000fffffffff, 0xfffffffffff00000},
		},

		// has zeros
		{
			Float128{0x3b2dffffffffffff, 0xffffffffffcffffe},
			Float128{0x8000008000000000, 0x000000003ffffffe},
			Float128{0x0000000000000000, 0x0000000000000000}, // 0
			Float128{0x8000000000000000, 0x0000000000000000}, // -0
		},
		{
			Float128{0x0000000000000000, 0x0000000000000000}, // +0
			Float128{0x0000000000000000, 0x0000000000000000}, // +0
			Float128{0x8000000000000000, 0x0000000000000000}, // -0
			Float128{0x0000000000000000, 0x0000000000000000}, // +0
		},
		{
			Float128{0x0000000000000000, 0x0000000000000000}, // +0
			Float128{0xc000000000000000, 0x0000000000000000}, // -2
			Float128{0x8000000000000000, 0x0000000000000000}, // -0
			Float128{0x8000000000000000, 0x0000000000000000}, // -0
		},

		// random tests
		{
			Float128{0x0002010000000000, 0x1fffffffffffffff},
			Float128{0x3e7b003fffbfffff, 0xffffffffffffffff},
			Float128{0x4080000000000200, 0x001fffffffffffff},
			Float128{0x4080000000000200, 0x001fffffffffffff},
		},
		{
			Float128{0x3ff7000000000000, 0x0000100000fffffe},
			Float128{0x7ffd000001ffffff, 0xfffffffffffffffe},
			Float128{0x3fffb272406353e8, 0x27de4059c093af48},
			Float128{0x7ff5000002000000, 0x00001000011ffffe},
		},
		{
			Float128{0x0002fff7ffffffff, 0xffffffffffffdfff},
			Float128{0xbffc0001ffffffff, 0xfffffffffff7ffff},
			Float128{0x403c000000000000, 0x0000000000000000},
			Float128{0x403c000000000000, 0x0000000000000000},
		},
		{
			Float128{0x407e000000000400, 0x0000000080000000},
			Float128{0xc0b8fffffff70000, 0x0000000000000000},
			Float128{0x4003000000000000, 0x00000000ffff0000},
			Float128{0xc137fffffff707ff, 0xffffdc00fffffffb},
		},
		{
			Float128{0x3ff2000000000000, 0x3fc0000000000000},
			Float128{0xc0dc000000ffffff, 0xfe00000000000000},
			Float128{0xc0b200000000f000, 0x0000000000000000},
			Float128{0xc0cf000001080000, 0x3dc7803fbfffff80},
		},
		{
			Float128{0xbf82000000000000, 0x0000000fc0000000},
			Float128{0xc000000000000000, 0x0000000000000000},
			Float128{0x3f812c9ce2963120, 0x9ef0dd1bd2c968f2},
			Float128{0x3f834b2738a58c48, 0x27bc3756b4b25a3c},
		},
		{
			Float128{0xbffbffffffefffff, 0xffffffffffffffff},
			Float128{0x4001ffffffffffff, 0xfffffffffffffc00},
			Float128{0x3ffeffffffefffff, 0xfffffffffffffbff},
			Float128{0xbf7b000000000000, 0x0000000010000000},
		},
	}
	for _, tt := range tests {
		got := FMA(tt.x, tt.y, tt.z)
		if !equals(got, tt.want) {
			t.Errorf("%s * %s + %s: got %s, want %s", dump(tt.x), dump(tt.y), dump(tt.z), dump(got), dump(tt.want))
		}
	}
}

func BenchmarkFMA(b *testing.B) {
	r := newXoshiro256pp()
	for i := 0; i < b.N; i++ {
		a, b := r.Float128Pair()
		c, _ := r.Float128Pair()
		runtime.KeepAlive(FMA(a, b, c))
	}
}
