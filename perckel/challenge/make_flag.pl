#!/usr/bin/perl

use strict;
use warnings;

use Storable qw(nstore);

no warnings 'once';

$Storable::Deparse = 1;

my $FLAG = 'CTF{ooohhh_purrrrl_<3}';

my  %f_hash = ();

sub encode_flag {
    my $f = shift;

    my $p = pack('D*', map{($_ * 1.0 ) + rand() * 1.0} unpack('C*', $f));

    return $p;
}


sub decode_flag {
    my $p = shift;

    my $f = pack('C*', map {int} unpack('D*', $p));

    return $f;
}


sub get_rand_string {

    my $s = '';
    my $l = int(rand() * 10.0) + 4;
    for (my $i = 0; $i < $l; $i++) {
        $s .= chr(ord('a') + int(rand() * 26.0));
    }

    return $s;
}


my $p = encode_flag($FLAG);

for (my $i = 0; $i < 1000; $i++) {
    $f_hash{get_rand_string()} = get_rand_string();
}

$f_hash{'flag'} = $p;
$f_hash{'decode_flag'} = \&decode_flag;

nstore(\%f_hash, 'flag.bin');
