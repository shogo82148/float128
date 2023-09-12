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
