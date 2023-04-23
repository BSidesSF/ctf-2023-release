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

my $flag = '';

my $o = $hdr - 1;
for (my $i = 0; $i < $h; $i++) {
    $o += $pad;
    my $p = substr($img, $o, 1);

    if (ord($p) != 0) {
        $flag .= $p;
    }
}

print 'flag: ', $flag, "\n";
