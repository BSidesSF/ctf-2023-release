#!/usr/bin/perl

my $h = 1024;
my $w = 1699; # gotta be odd to have the padding

my $img;
open (my $IN, '<', $ARGV[0]) or die 'Unable to open file: ', $!, "\n";
{
    local $/ = undef;
    $img  = <$IN>;
}
close $IN;

my $hdr = 32;
my $pad = $w * 3 + 1;

my $flag = 'CTF{black_padding_information_paradox}';

my $o = $hdr - 1;
foreach my $l (split(//, $flag)) {
    $o += $pad;
    substr($img, $o, 1) = $l;
}

open (my $OUT, '>', $ARGV[1]) or die 'Unable to open file: ', $!, "\n";
print $OUT $img;
close $OUT;
