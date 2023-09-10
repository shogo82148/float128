#!/bin/env perl

use v5.38;
use utf8;
use strict;
use FindBin;

sub format_uint128($v) {
    unless ($v =~ /^([0-9a-fA-F]{16})([0-9a-fA-F]{16})$/) {
        die "Invalid uint128: $v";
    }
    my ($h, $l) = ($1, $2);
    return sprintf("Float128{0x%s, 0x%s}", $h, $l);
}

# f128_lt.txt is generated by TestFloat-3b/testfloat_gen.
# http://www.jhauser.us/arithmetic/TestFloat.html
# $ ./testfloat_gen -level 1 f128_lt > f64_to_f16.txt
open my $fh, "<", "$FindBin::Bin/f128_lt.txt" or die "Can't open f128_lt.txt: $!";

say <<EOF;
// Code generated by scripts/f128_lt.pl; DO NOT EDIT.

package float128

var f128Lt = []struct {
    a, b Float128
    want bool
} {
EOF
while(my $line = <$fh>) {
    chomp $line;
    my ($a, $b, $c, undef) = split /\s+/, $line;
    printf("{%s, %s, %s},\n", format_uint128($a), format_uint128($b), $c ne "0" ? "true" : "false");
}

say "}";

close $fh;
