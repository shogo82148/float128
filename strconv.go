package float128

import (
	"github.com/shogo82148/int128"
)

func (x Float128) Text(fmt byte, prec int) string {
	return string(x.Append(nil, fmt, prec))
}

func (x Float128) Append(buf []byte, fmt byte, prec int) []byte {
	switch {
	case x.IsNaN():
		return append(buf, "NaN"...)
	case x == inf:
		return append(buf, "+Inf"...)
	case x == neginf:
		return append(buf, "-Inf"...)
	}

	switch fmt {
	case 'b':
		return x.appendBin(buf)
	case 'x', 'X':
		return x.appendHex(buf, fmt, prec)
	}

	return buf
}

func (x Float128) appendBin(buf []byte) []byte {
	if x.h&signMask128H != 0 {
		buf = append(buf, '-')
	}
	exp := int32((x.h>>(shift128-64))&mask128) - bias128
	frac := int128.Uint128{H: x.h & fracMask128H, L: x.l}
	if exp == -bias128 {
		exp++
	} else {
		frac.H |= 1 << (shift128 - 64)
	}
	exp -= shift128

	buf = frac.Append(buf, 10)
	buf = append(buf, 'p')
	if exp >= 0 {
		buf = append(buf, '+')
	} else {
		buf = append(buf, '-')
		exp = -exp
	}
	switch {
	case exp >= 10000:
		buf = append(buf, '0'+byte(exp/10000))
		exp %= 10000
		fallthrough
	case exp >= 1000:
		buf = append(buf, '0'+byte(exp/1000))
		exp %= 1000
		fallthrough
	case exp >= 100:
		buf = append(buf, '0'+byte(exp/100))
		exp %= 100
		fallthrough
	case exp >= 10:
		buf = append(buf, '0'+byte(exp/10))
		exp %= 10
		fallthrough
	default:
		buf = append(buf, '0'+byte(exp))
	}
	return buf
}

func (x Float128) appendHex(buf []byte, fmt byte, prec int) []byte {
	sign, exp, frac := x.split()
	if sign != 0 {
		buf = append(buf, '-')
	}
	buf = append(buf, '0', fmt)
	if frac == (int128.Uint128{}) {
		buf = append(buf, '0')
		if prec > 0 {
			buf = append(buf, '.')
			for i := 0; i < prec; i++ {
				buf = append(buf, '0')
			}
		}
		buf = append(buf, 'P'|(fmt&0x20), '+', '0', '0')
		return buf
	}

	if prec < 0 {
		var tmp [30]byte
		digits := frac.Append(tmp[:0], 16)

		// find the last non-zero digit
		n := len(digits) - 1
		for ; n >= 0 && digits[n] == '0'; n-- {
		}

		// print the digits
		buf = append(buf, '1')
		if n > 0 {
			buf = append(buf, '.')
			buf = append(buf, digits[1:n+1]...)
		}
	} else if prec < 28 {
		// round nearest even
		shift := uint(shift128 - 4*prec)
		one := int128.Uint128{L: 1}
		offset := one.Lsh(shift - 1).Sub(one)
		frac = frac.Add(offset).Add(frac.Rsh(shift).And(one))
		if frac.H >= 2<<(shift128-64) {
			exp++
			frac = frac.Rsh(1)
		}

		// print the digits
		buf = append(buf, '1')
		if prec > 0 {
			buf = append(buf, '.')
			var tmp [30]byte
			digits := frac.Append(tmp[:0], 16)
			buf = append(buf, digits[1:prec+1]...)
		}
	} else {
		var tmp [30]byte
		digits := frac.Append(tmp[:0], 16)
		buf = append(buf, '1', '.')
		buf = append(buf, digits[1:29]...)
		for i := 28; i < prec; i++ {
			buf = append(buf, '0')
		}
	}

	buf = append(buf, 'P'|(fmt&0x20))
	if exp >= 0 {
		buf = append(buf, '+')
	} else {
		buf = append(buf, '-')
		exp = -exp
	}

	switch {
	case exp >= 10000:
		buf = append(buf, '0'+byte(exp/10000))
		exp %= 10000
		fallthrough
	case exp >= 1000:
		buf = append(buf, '0'+byte(exp/1000))
		exp %= 1000
		fallthrough
	case exp >= 100:
		buf = append(buf, '0'+byte(exp/100))
		exp %= 100
		fallthrough
	default:
		buf = append(buf, byte((exp/10)%10)+'0', byte(exp%10)+'0')
	}
	return buf
}
