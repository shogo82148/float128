package float128

import (
	"fmt"
	"math"
	"math/bits"

	"github.com/shogo82148/int128"
)

var nan = Float128{0x7fff_8000_0000_0000, 0x00}
var inf = Float128{0x7fff_0000_0000_0000, 0x00}
var neginf = Float128{0xffff_0000_0000_0000, 0x00}
var one = int128.Uint128{L: 1}

const (
	mask128      = 0x7fff       // mask for exponent
	shift128     = 128 - 15 - 1 // shift for exponent
	bias128      = 16383        // bias for exponent
	signMask128H = 1 << 63      // mask for sign bit
	fracMask128H = 1<<(shift128-64) - 1
	qNaNBitH     = (1 << (shift128 - 64 - 1))
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
			return Float128{
				sign | (mask128 << (shift128 - 64)) | (frac >> (64 - shift128 + shift64)) | qNaNBitH,
				frac << (shift128 - shift64),
			}
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

// Float64 returns the float64 representation of f.
func (f Float128) Float64() float64 {
	sign := f.h & signMask128H
	exp := int((f.h >> (shift128 - 64)) & mask128)
	frac := int128.Uint128{H: f.h & fracMask128H, L: f.l}

	if exp == mask128 {
		if frac.H|frac.L != 0 {
			// f is NaN
			const qNaNBit = 1 << (shift64 - 1)
			f64 := sign | (mask64 << shift64) | qNaNBit | (frac.H << (64 - shift128 + shift64)) | (frac.L >> (shift128 - shift64))
			return math.Float64frombits(f64)
		} else {
			// f is ±Inf
			return math.Float64frombits(sign | (mask64 << shift64))
		}
	}

	exp -= bias128
	if exp <= -bias64 {
		roundBit := -exp + shift128 - (bias64 + shift64 - 1)
		frac.H |= 1 << (shift128 - 64)
		frac = frac.Add(one.Lsh(uint(roundBit - 1)).Sub(one))
		frac = frac.Add(frac.Rsh(uint(roundBit)).And(one))
		frac = frac.Rsh(uint(roundBit))
		return math.Float64frombits(sign | frac.L)
	}
	if exp >= mask64-bias64 {
		// overflow, the result is ±Inf
		return math.Float64frombits(sign | (mask64 << shift64))
	}

	// round to nearest, tie to even
	var carry uint64
	frac.L, carry = bits.Add64(frac.L, 0x7ff_ffff_ffff_ffff+(frac.L>>(shift128-shift64)&1), 0)
	frac.H, _ = bits.Add64(f.h, 0, carry)
	exp = int((frac.H>>(shift128-64))&mask128) - mask128 + bias64
	frac.H &= fracMask128H

	if exp >= mask64 {
		// overflow, the result is ±Inf
		return math.Float64frombits(sign | (mask64 << shift64))
	}

	frac = frac.Rsh(shift128 - shift64)
	return math.Float64frombits(sign | uint64(exp)<<shift64 | frac.L)
}

func (f Float128) GoString() string {
	sign := f.h & signMask128H
	c := '+'
	if sign != 0 {
		c = '-'
	}
	exp := int((f.h >> (shift128 - 64)) & mask128)
	frac := int128.Uint128{H: f.h & fracMask128H, L: f.l}

	if exp == mask128 {
		if frac.H|frac.L != 0 {
			// f is NaN
			return "NaN"
		} else {
			// f is ±Inf
			return fmt.Sprintf("%cInf", c)
		}
	} else if exp == 0 {
		if frac.H|frac.L == 0 {
			return fmt.Sprintf("%c0x0.0000000000000000000000000000p+0", c)
		}
		return fmt.Sprintf("%c0x0.%012x%016xp%+d", c, frac.H, frac.L, -bias128+1)
	}
	return fmt.Sprintf("%c0x1.%012x%016xp%+d", c, frac.H, frac.L, exp-bias128)
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

func (f Float128) isSignalingNaN() bool {
	const expMask = (mask128 << (shift128 - 64))
	return f.h&expMask == expMask && f.h&fracMask128H != 0 && f.l != 0 && f.h&qNaNBitH == 0
}

func (f Float128) isQuietNaN() bool {
	const expMask = (mask128 << (shift128 - 64))
	return f.h&expMask == expMask && f.h&fracMask128H != 0 && f.l != 0 && f.h&qNaNBitH != 0
}

func propagateNaN(a, b Float128) Float128 {
	if a.isSignalingNaN() {
		return Float128{a.h | qNaNBitH, a.l}
	}
	if b.isSignalingNaN() {
		return Float128{b.h | qNaNBitH, b.l}
	}
	if a.isQuietNaN() {
		return a
	}
	if b.isQuietNaN() {
		return b
	}
	panic("never reach here")
}
