package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/shogo82148/float128"
)

func main() {
	switch os.Args[1] {
	case "f128_to_f64":
		f128_to_f64()
	case "f64_to_f128":
		f64_to_f128()
	case "f128_mul":
		f128_mul()
	case "f128_div":
		f128_div()
	case "f128_add":
		f128_add()
	case "f128_eq":
		f128_eq()
	case "f128_le":
		f128_le()
	case "f128_lt":
		f128_lt()
	case "f128_mulAdd":
		f128_mulAdd()
	}
}

func f128_to_f64() {
	var failed int64
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		line := s.Text()
		line = strings.TrimSpace(line)

		// the input is a 128-bit floating point number.
		s128, line, _ := strings.Cut(line, " ")
		f128, err := parseFloat128(s128)
		if err != nil {
			log.Fatal(err)
		}

		// the output is a 64-bit floating point number.
		s64, line, _ := strings.Cut(line, " ")
		f64, err := strconv.ParseUint(s64, 16, 64)
		if err != nil {
			log.Fatal(err)
		}

		// test converting
		got := math.Float64bits(f128.Float64())
		if got != f64 {
			fmt.Printf("%s %s %016x\n", s128, s64, got)
			failed++
		}
		_ = line
	}
	if failed > 0 {
		fmt.Printf("%d tests failed\n", failed)
		os.Exit(1)
	}
}

func f64_to_f128() {
	var failed int64
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		line := s.Text()
		line = strings.TrimSpace(line)

		// the input is a 64-bit floating point number.
		s64, line, _ := strings.Cut(line, " ")
		b64, err := strconv.ParseUint(s64, 16, 64)
		if err != nil {
			log.Fatal(err)
		}
		f64 := math.Float64frombits(b64)

		// the output is a 128-bit floating point number.
		s128, line, _ := strings.Cut(line, " ")
		f128, err := parseFloat128(s128)
		if err != nil {
			log.Fatal(err)
		}

		// test converting
		h0, l0 := f128.Bits()
		h1, l1 := float128.FromFloat64(f64).Bits()
		if h0 != h1 || l0 != l1 {
			fmt.Printf("%s %s %08x%08x\n", s64, s128, h1, l1)
			failed++
		}
		_ = line
	}
	if failed > 0 {
		fmt.Printf("%d tests failed\n", failed)
		os.Exit(1)
	}
}

func f128_mul() {
	var failed int64
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		line := s.Text()
		line = strings.TrimSpace(line)

		a, b, c, err := parseFloat128x3(line)
		if err != nil {
			log.Fatal(err)
		}

		got := a.Mul(b)
		if got.IsNaN() && c.IsNaN() {
			continue
		}
		if got != c {
			fmt.Printf("%s %s %s %s\n", dump(a), dump(b), dump(c), dump(got))
			failed++
		}
	}
	if failed > 0 {
		fmt.Printf("%d tests failed\n", failed)
		os.Exit(1)
	}
}

func f128_div() {
	var failed int64
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		line := s.Text()
		line = strings.TrimSpace(line)

		a, b, c, err := parseFloat128x3(line)
		if err != nil {
			log.Fatal(err)
		}

		got := a.Quo(b)
		if got.IsNaN() && c.IsNaN() {
			continue
		}
		if got != c {
			fmt.Printf("%s %s %s %s\n", dump(a), dump(b), dump(c), dump(got))
			failed++
		}
	}
	if failed > 0 {
		fmt.Printf("%d tests failed\n", failed)
		os.Exit(1)
	}
}

func f128_add() {
	var failed int64
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		line := s.Text()
		line = strings.TrimSpace(line)

		a, b, c, err := parseFloat128x3(line)
		if err != nil {
			log.Fatal(err)
		}

		got := a.Add(b)
		if got.IsNaN() && c.IsNaN() {
			continue
		}
		if got != c {
			fmt.Printf("%s %s %s %s\n", dump(a), dump(b), dump(c), dump(got))
			failed++
		}
	}
	if failed > 0 {
		fmt.Printf("%d tests failed\n", failed)
		os.Exit(1)
	}
}

func f128_eq() {
	var failed int64
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		line := s.Text()
		line = strings.TrimSpace(line)

		a, b, c, err := parseFloat128x2Bool(line)
		if err != nil {
			log.Fatal(err)
		}

		got := a.Eq(b)
		if got != c {
			fmt.Printf("%s %s %t %t\n", dump(a), dump(b), c, got)
			failed++
		}
	}
	if failed > 0 {
		fmt.Printf("%d tests failed\n", failed)
		os.Exit(1)
	}
}

func f128_le() {
	var failed int64
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		line := s.Text()
		line = strings.TrimSpace(line)

		a, b, c, err := parseFloat128x2Bool(line)
		if err != nil {
			log.Fatal(err)
		}

		got := a.Le(b)
		if got != c {
			fmt.Printf("%s %s %t %t\n", dump(a), dump(b), c, got)
			failed++
		}
	}
	if failed > 0 {
		fmt.Printf("%d tests failed\n", failed)
		os.Exit(1)
	}
}

func f128_lt() {
	var failed int64
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		line := s.Text()
		line = strings.TrimSpace(line)

		a, b, c, err := parseFloat128x2Bool(line)
		if err != nil {
			log.Fatal(err)
		}

		got := a.Lt(b)
		if got != c {
			fmt.Printf("%s %s %t %t\n", dump(a), dump(b), c, got)
			failed++
		}
	}
	if failed > 0 {
		fmt.Printf("%d tests failed\n", failed)
		os.Exit(1)
	}
}

func f128_mulAdd() {
	var failed int64
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		line := s.Text()
		line = strings.TrimSpace(line)

		a, b, c, d, err := parseFloat128x4(line)
		if err != nil {
			log.Fatal(err)
		}

		got := float128.FMA(a, b, c)
		if got.IsNaN() && c.IsNaN() {
			continue
		}
		if got != d {
			fmt.Printf("%s %s %s %s %s\n", dump(a), dump(b), dump(c), dump(d), dump(got))
			failed++
		}
	}
	if failed > 0 {
		fmt.Printf("%d tests failed\n", failed)
		os.Exit(1)
	}
}

func parseFloat128(s string) (float128.Float128, error) {
	if len(s) != 32 {
		return float128.Float128{}, fmt.Errorf("invalid length: %d", len(s))
	}
	h, err := strconv.ParseUint(s[:16], 16, 64)
	if err != nil {
		return float128.Float128{}, err
	}
	l, err := strconv.ParseUint(s[16:], 16, 64)
	if err != nil {
		return float128.Float128{}, err
	}

	return float128.FromBits(h, l), nil
}

func parseFloat128x3(s string) (a, b, c float128.Float128, err error) {
	sa, s, _ := strings.Cut(s, " ")
	sb, s, _ := strings.Cut(s, " ")
	sc, s, _ := strings.Cut(s, " ")
	_ = s
	a, err = parseFloat128(sa)
	if err != nil {
		return
	}
	b, err = parseFloat128(sb)
	if err != nil {
		return
	}
	c, err = parseFloat128(sc)
	if err != nil {
		return
	}
	return
}

func parseFloat128x4(s string) (a, b, c, d float128.Float128, err error) {
	sa, s, _ := strings.Cut(s, " ")
	sb, s, _ := strings.Cut(s, " ")
	sc, s, _ := strings.Cut(s, " ")
	sd, s, _ := strings.Cut(s, " ")
	_ = s
	a, err = parseFloat128(sa)
	if err != nil {
		return
	}
	b, err = parseFloat128(sb)
	if err != nil {
		return
	}
	c, err = parseFloat128(sc)
	if err != nil {
		return
	}
	d, err = parseFloat128(sd)
	if err != nil {
		return
	}
	return
}

func parseFloat128x2Bool(s string) (a, b float128.Float128, c bool, err error) {
	sa, s, _ := strings.Cut(s, " ")
	sb, s, _ := strings.Cut(s, " ")
	sc, s, _ := strings.Cut(s, " ")
	_ = s
	a, err = parseFloat128(sa)
	if err != nil {
		return
	}
	b, err = parseFloat128(sb)
	if err != nil {
		return
	}
	c = sc != "0"
	return
}

func dump(f float128.Float128) string {
	h, l := f.Bits()
	return fmt.Sprintf("%016x%016x", h, l)
}
