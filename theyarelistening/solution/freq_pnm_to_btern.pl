#!/usr/bin/perl

use strict;
use warnings;


my @row = ();
push @row, [(0) x 300];
push @row, [(0) x 300];
push @row, [(0) x 300];

my $lnum = 0;
my $col = 0;
while (<STDIN>) {
    my $line = $_;
    chomp($line);
    $lnum++;

    next if ($lnum < 4);

    my @cols = split(/\s+/);
    foreach my $c (@cols) {
        if ($c == 0) {
            $row[int($col / 300)][$col % 300] = 1;
        }
        $col++;
    }
}


for (my $i = 0; $i < 300; $i++) {
    my $v = 2 * $row[2][$i] + 1 * $row[0][$i];

    print $v;
}

print "\n";
