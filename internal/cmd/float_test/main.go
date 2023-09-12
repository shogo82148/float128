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
	}
}

func f128_to_f64() {
	var failed int64
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		line := s.Text()
		line = strings.TrimSpace(line)
		s128, line, _ := strings.Cut(line, " ")
		f128, err := parseFloat128(s128)
		if err != nil {
			log.Fatal(err)
		}
		s64, _, _ := strings.Cut(line, " ")
		f64, err := strconv.ParseUint(s64, 16, 64)
		if err != nil {
			log.Fatal(err)
		}
		got := math.Float64bits(f128.Float64())
		if got != f64 {
			fmt.Printf("%s %s %016x\n", s128, s64, got)
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
