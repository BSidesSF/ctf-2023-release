#!/usr/bin/perl

use strict;
use warnings;

no warnings 'once';

use Storable qw(retrieve);

$Storable::Eval = 1;


my $f_ref = retrieve('flag.bin');

print $f_ref->{'decode_flag'}($f_ref->{'flag'}), "\n";
