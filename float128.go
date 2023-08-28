package float128

import (
	"math"
	"math/bits"
)

var nan = Float128{0x7fff_8000_0000_0000, 0x01}
var inf = Float128{0x7fff_0000_0000_0000, 0x00}
var neginf = Float128{0xffff_0000_0000_0000, 0x00}

const (
	mask128      = 0x7fff       // mask for exponent
	shift128     = 128 - 15 - 1 // shift for exponent
	bias128      = 16383        // bias for exponent
	signMask128H = 1 << 63      // mask for sign bit
	fracMask128H = 1<<(shift128-64) - 1
)

const (
	mask64     = 0x7ff       // mask for exponent
	shift64    = 64 - 11 - 1 // shift for exponent
	bias64     = 1023        // bias for exponent
	signMask64 = 1 << 63     // mask for sign bit
	fracMask64 = 1<<shift64 - 1
)

// Float128 represents a 128-bit floating point number.
type Float128 struct {
	h, l uint64
}

// NaN returns a Float128 representation of NaN.
func NaN() Float128 {
	return nan
}

// IsNaN reports whether f is NaN.
func (f Float128) IsNaN() bool {
	const expMask = (mask128 << (shift128 - 64))
	return f.h&expMask == expMask && f.h&fracMask128H != 0 && f.l != 0
}

// Inf returns positive infinity if sign >= 0, negative infinity if sign < 0.
func Inf(sign int) Float128 {
	if sign >= 0 {
		return inf
	} else {
		return neginf
	}
}

// IsInf reports whether f is an infinity, according to sign.
// If sign > 0, IsInf reports whether f is positive infinity.
// If sign < 0, IsInf reports whether f is negative infinity.
// If sign == 0, IsInf reports whether f is either infinity.
func (f Float128) IsInf(sign int) bool {
	return sign >= 0 && f == inf || sign <= 0 && f == neginf
}

// Bits returns the IEEE 754 binary representation of f.
func (f Float128) Bits() (h, l uint64) {
	return f.h, f.l
}

// FromFloat128 returns the floating point number corresponding
// to the IEEE 754 binary representation of f.
func FromFloat64(f float64) Float128 {
	b := math.Float64bits(f)
	sign := b & signMask64
	exp := int((b >> shift64) & mask64)
	frac := b & fracMask64

	if exp == mask64 {
		if frac != 0 {
			// f is NaN
			return nan
		} else {
			// f is ±Inf
			return Float128{
				inf.h | sign,
				inf.l,
			}
		}
	}

	if exp == 0 {
		// f is subnormal
		if frac == 0 {
			return Float128{sign, 0}
		}

		// normalize f
		l := bits.Len64(frac)
		exp = l - shift64
		frac = (frac << (53 - l)) & fracMask64
	}

	exp += bias128 - bias64
	return Float128{
		sign | uint64(exp)<<(shift128-64) | (frac >> (64 - shift128 + shift64)),
		frac << (shift128 - shift64),
	}
}
