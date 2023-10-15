package float128

import "github.com/shogo82148/int128"

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
