#!/usr/bin/perl

use strict;
use warnings;

use Math::Trig ':pi';

my $keybits = 120;

# perl -e 'for (my $i = 0; $i < 256; $i++) { print sprintf("%02x", $i); } print "\n";'
#000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f404142434445464748494a4b4c4d4e4f505152535455565758595a5b5c5d5e5f606162636465666768696a6b6c6d6e6f707172737475767778797a7b7c7d7e7f808182838485868788898a8b8c8d8e8f909192939495969798999a9b9c9d9e9fa0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c8c9cacbcccdcecfd0d1d2d3d4d5d6d7d8d9dadbdcdddedfe0e1e2e3e4e5e6e7e8e9eaebecedeeeff0f1f2f3f4f5f6f7f8f9fafbfcfdfeff

my $enc = "fffffce6f3e5e5c9d3d1bf87aa94969384e3768a6c1765a462d96300652a68506b7b6de46f0c6ec86d436aee686b667365b166ac69ae6eb7757d7d7785e88e0094f39a129ce39d309b0996bf90dc8a0982ff7c6a76da72af701a6f176f7370df72f8755e77bb79d17b827cc87db97e787f2d7ffb80f78224837184ba85d4869086c5865a8547839a81757f047c7d7a0f77de75fe746c731371ce70726ed46cd76a7167b064c061e05f605d965ccf5d455f17623e66926bc97182774f7cc5818b8561882c89f48ae78b4b8b768bbc8c658d9e8f7291ca946a96ff99269a809abb99a39727935f8e8788fa83247d767855740f70d36eaa6d7d6d176d346d8b6ddc6dfc6ddd6d8c6d356d186d7e6eab70d4741078567d75832588fb8e889360972899a49abc9a8199279700946b91cb8f738d9f8c668bbd8b778b4c8ae889f5882d8562818c7cc6775071836bca6693623f5f185d465cd05d975f6161e164c167b16a726cd86ed5707371cf7314746d75ff77df7a107c7e7f058176839b8548865b86c6869185d584bb8372822580f87ffc7f2e7e797dba7cc97b8379d277bc755f72f970e06f746f18701b72b076db7c6b83008a0a90dd96c09b0a9d319ce49a1394f48e0185e97d78757e6eb869af66ad65b26674686c6aef6d446ec96f0d6de56b7c6851652b630162da65a56c18768b84e49694aa95bf88d3d2e5caf3e6fce7";

my @enc_vals = unpack('n*', pack('H*', $enc));


my @dct = ();
my $dct_sum = 0;
for (my $f = 0; $f < 16; $f++) {
    #my $dct_f = dct_freq($f, \@scaled_yval);
    my $dct_f = dct_freq($f, \@enc_vals);

    push @dct, $dct_f;

    $dct_sum += $dct_f;
}

my $klen = $keybits / 4;
my $expected_dct_sum = (($klen * ($klen - 1)) / 2) * 256.0 + ($klen * 32.0) + ($klen / 3) * 4;
my $dct_scale = ($expected_dct_sum * 1.0) / ($dct_sum * 1.0);

print sprintf("Key bits %d; expected dct sum: %.0f; dct scale factor: %.03f\n", $klen * 4, $expected_dct_sum, $dct_scale);

my @sums = ();
my @copies = ();
my @mod = ();

print '==dct table==', "\n";
for (my $f = 0; $f < 16; $f++) {
    #my $dct_f = dct_freq($f, \@scaled_yval);
    my $dct_f = dct_freq($f, \@enc_vals);


    my $coeff = int(sprintf('%.0f', ($dct_f * $dct_scale) / 4.0)) * 4;
    my $n = int(($coeff % 256) / 32);
    my $s = int(($coeff - ($n * 32)) / 256);
    my $m = int((($coeff - ($n * 32)) - ($s * 256)) / 4);

    warn sprintf('%-2d: %.03f (rescaled coeff: %d; copies: %d; pos sum: %d; mod copies: %d)', $f, $dct_f, $coeff, $n, $s, $m), "\n";
    push @sums, $s;
    push @copies, $n;
    push @mod, $m;
}

my @all = (0 .. (($keybits / 4) - 1));
#my @found = ();
my @running = ();

#find_sums($copies[2], $sums[2], 0, \@all, \@running, \@found);
#foreach my $sol (@found) {
#    print 'found: ', join(', ', @{$sol}), "\n";
#}

solve(0, \@all, \@running);

sub solve {
    my $i = shift; # Which we're on
    my $avref = shift; # available
    my $lref = shift; # running list

    if ($i == 16) {

        my $key = " " x ($keybits / 4);

        for (my $k = 0; $k < 16; $k++) {
            my $a = sprintf('%01x', $k);
            foreach my $p (@{$lref->[$k]}) {
                substr($key, $p, 1) = $a;
            }
        }

        print sprintf('candidate key %s (%s)', $key, pack('H*', $key)), "\n";
        return;
    }


    my @av = @{$avref};
    my @l = ();

    if ($copies[$i] == 0) {
        my @l = @{$lref};
        push @l, [()];

        solve($i + 1, \@av, \@l)
    } else {
        my @found = ();
        my @running = ();
        #warn 'av list: ', join(', ', @av), "\n";
        find_sums($copies[$i], $sums[$i], 0, 0, $mod[$i], \@av, \@running, \@found);

        if (scalar @found == 0) {
            return
        }

        foreach my $sol (@found) {
            my @l = @{$lref};

            my @s = @{$sol};

            push @l, \@s;
            my $nav = subtract_set(\@av, \@s);

            solve($i + 1, $nav, \@l);
        }
    }
}


sub find_sums {
    my $n = shift; # How many
    my $g = shift; # goal sum
    my $s = shift; # runing sum
    my $e = shift; # number of mod
    my $ge = shift; # goal number of mod
    my $avref = shift; # avail list
    my $lref = shift; # running list
    my $solref = shift; # found solutions

    if ($s > $g) {
        return;
    }

    if (($n == 0) && ($s < $g)) {
        return;
    }

    if ($n > scalar(@$avref)) {
        return;
    }

    if (($n == 0) && ($e != $ge)) {
        return;
    }

    if (($n == 0) && ($s == $g) && ($e == $ge)) {
        # we have a solution
        push @$solref, [@{$lref}];
        return;
    }

    my @av = @{$avref};
    my @l = @{$lref};

    my $v = shift @av;
    push @l, $v;

    my $eo = 0;
    if ($v % 3 == 0) {
        $eo = 1;
    }

    unless (defined $v) {
        warn 'undef v, passed av list: ', join(', ', @{$avref}), "\n";
        warn sprintf('n: %d; g: %d; s: %d', $n, $g, $s), "\n";

        if ($n == 0) {
            warn 'n is zero', "\n";
        } else {
            warn 'n is NOT zero', "\n";
        }

        if (($n == 0) && ($s < $g)) {
            warn 'test passed', "\n";
        }
    }

    find_sums($n - 1, $g, $s + $v, $e + $eo, $ge, \@av, \@l, $solref);

    @l = @{$lref};
    find_sums($n, $g, $s, $e, $ge, \@av, \@l, $solref);
}


sub subtract_set {
    my $all = shift;
    my $mref = shift;

    my %m = ();

    foreach my $v (@{$mref}) {
        $m{$v} = 1;
    }

    my @new = ();

    foreach my $v (@{$all}) {
        push @new, $v unless (exists $m{$v});
    }

    return \@new;
}


sub dct_freq {
    my $kfreq = shift; # key nibble freq 0 to 15
    my $yval_ref = shift;

    die unless (scalar(@{$yval_ref}) == 256);

    my $dct_f = 0.0;

    for (my $x = 0; $x < 256; $x++) {
        $dct_f += $yval_ref->[$x] * cos(((2.0 * pi) / 256.0) * ($kfreq + 1.0) * $x);
    }

    return $dct_f / 128.0;
}
