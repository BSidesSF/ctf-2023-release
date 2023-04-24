#!/usr/bin/perl

use strict;
use warnings;

use bignum;

my $flag = 'perasperaadastra';

print 'num: ', key_to_number($flag), "\n";

sub key_to_number {
    my $w = shift;

    return -1 unless ($w =~ m/^[a-z]+$/);

    my @lets = split(//, $w);

    my $n = 0;

    foreach my $l (@lets) {
        my $c = 0;
        if ($l =~ m/^[a-z]$/) {
            $c = (ord($l) - ord('a')) + 1;
        }

        # base 27
        $n = ($n * 27) + $c;
    }

    #print 'word ', $w, ' -> ', $n, "\n";

    return $n;
}
