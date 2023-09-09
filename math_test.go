package float128

import (
	"fmt"
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
		if got != tt.want {
			t.Errorf("%s * %s: got %s, want %s", dump(tt.a), dump(tt.b), dump(got), dump(tt.want))
		}
	}
}

//go:generate sh -c "perl scripts/f128_mul.pl | gofmt > f128_mul_test.go"

func TestMul_TestFloat(t *testing.T) {
	for _, tt := range f128Mul {
		tt := tt
		fa := tt.a
		fb := tt.b
		func() {
			defer func() {
				err := recover()
				if err != nil {
					t.Errorf("%s * %s: want %s, panic %#v", dump(fa), dump(fb), dump(tt.want), err)
				}
			}()
			got := fa.Mul(fb)
			if got != tt.want {
				t.Errorf("%s * %s: got %s, want %s", dump(fa), dump(fb), dump(got), dump(tt.want))
			}
		}()
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

		// 0 + anything => anything
		{
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000},
			Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1
			Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1
		},
	}

	for _, tt := range tests {
		got := tt.a.Add(tt.b)
		if got != tt.want {
			t.Errorf("%s + %s: got %s, want %s", dump(tt.a), dump(tt.b), dump(got), dump(tt.want))
		}
	}
}

//go:generate sh -c "perl scripts/f128_add.pl | gofmt > f128_add_test.go"

func TestAdd_TestFloat(t *testing.T) {
	for _, tt := range f128Add {
		tt := tt
		fa := tt.a
		fb := tt.b
		func() {
			defer func() {
				err := recover()
				if err != nil {
					t.Errorf("%#v + %#v: want %#v, panic! %v", fa, fb, tt.want, err)
				}
			}()
			got := fa.Add(fb)
			if got != tt.want {
				t.Errorf("%#v (%016x_%016x) + %#v (%016x_%016x): got %#v (%016x_%016x), want %#v (%016x_%016x)",
					fa, fa.h, fa.l, fb, fb.h, fb.l, got, got.h, got.l, tt.want, tt.want.h, tt.want.l)
			}
		}()
	}
}
