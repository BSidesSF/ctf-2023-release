#!/usr/bin/perl

use strict;
use warnings;

use MIME::Base64;
use Math::Trig ':pi';

use bignum upgrade => undef; # Keep things as integers

my $W = 950;
my $H = 950;
my $R = 0.60; # Percent of a ring's radius
my $OUT_W = 1500; # Everything scaled to this
my $MAX_COLS = 4;
my $rings = 6;

my $IDIR = './imgs/grey';

my $SNAME = 'symmetric\'s truly horrible perl hack to approximate a webserver';

my $HTML_START = << 'ENDHTMLSTART';
<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="content-type" content="text/html; charset=UTF-8">
    <title>alien</title>
    <style type="text/css">
      div {
          padding-top: 16px;
          padding-right: 16px;
          padding-bottom: 16px;
          padding-left: 16px;
      }
      .container {
          align:center;
      }
      .alien {
          float:center;
          text-align:center;
      }
      .form {
          color: white;
          font-size: 16pt;
      }
      .back {
          background-image: repeating-linear-gradient(rgba(25, 25, 25, 1), rgba(50, 50, 50), rgba(25, 25, 25, 1));
      }
      img {
          border: 5px solid #323232;
      }
    </style>
  </head>
  <body class="back">
    <div class="form">
         <form method="get" action="/">
               <p>Clusters:<br /><input type="text" name="clusters" value="%s" size="8" maxlength="8" /></p>
               <p>Sequence:<br /><input type="text" name="seq" value="%s" size="8" maxlength="8" /></p>
               <p>Binary:<br /><input type="text" name="bin" value="%s" size="100" /></p>
               <p><input type="submit" name="click" value="Transmit" /></p>
        </form>
    </div>
    <div class="container">
ENDHTMLSTART
;

my $HTML_END = << 'ENDHTMLEND';
    </div>
  </body>
</html>
ENDHTMLEND
;


#alarm 5;


my $lc = 0;
my $got_get = 0;
my $error = 0;
my $hc = 0;
my $clusters = '';
my $seq = '';
my $bin = '';
while (<STDIN>) {
    $lc++;

    my $line = $_;

    if ($lc == 1) {
        #warn 'line 1', "\n";
        #warn $line, "\n";
        if ($line =~ m/^GET\s\/\?.*?\sHTTP\/1\.[01]\s+$/) {
            $got_get = 1;
            if ($line =~ m/clusters=([1-6]+)/) {
                $clusters = $1
            } else {
                $got_get = 0;
            }

            if ($line =~ m/bin=([01]+)/) {
                $bin = $1
            } else {
                $got_get = 0;
            }

            if ($line =~ m/seq=([2-8]+)/) {
                $seq = $1
            } else {
                $got_get = 0;
            }

            #warn 'got get with key ', $key, ' and msg ', $msg, "\n";
        }
    } else {
        if ($line =~ m/^User-Agent: .*GoogleHC.*$/) {
            $hc = 1;
        }
        last if ($line =~ m/^[\r\n]+$/);
    }
}

if ($hc == 1) {
  print "HTTP/1.0 200 OK\r\n";
  print "\r\n";
  exit 0;
}


$error = 1 unless ($got_get == 1);


#warn 'Converting bin ', $bin, ' to num', "\n";
my $n = bin_to_num($bin);

$error = 1 if ($n <= 0);
#warn 'Number: ', $n, "\n";

my $encref;
if ($error == 0) {
    #warn 'Encoding number with clusters ', $clusters, ' and seq ', $seq, "\n";
    $encref = num_to_enc($n, $clusters, $seq);

    $error = 1 if (scalar(@{$encref}) < 1); # should be impossible
    $error = 1 if (scalar(@{$encref}) > 60);

    my $rcount = 0;
    foreach my $c (@{$encref}) {
        $rcount += scalar(@{$c});
        #warn 'cluster: [', join(', ', @{$c}), ']', "\n";
    }

    $error = 1 if ($rcount > 200); # too many rings
}

if ($error == 1) {
    print 'HTTP/1.0 303 See Other', "\r\n";
    print "Server: ", $SNAME, "\r\n";
    print 'Location: /?clusters=123456&seq=2345678&bin=0110100001100101011011000110110001101111', "\r\n";
    print "\r\n";

    exit 0;
}

my $html = sprintf($HTML_START, $clusters, $seq, $bin);

$html .= "\n" . img_to_html(enc_to_img($encref)) . "\n";

$html .= $HTML_END;

print 'HTTP/1.0 200 OK', "\r\n";
print 'Server: ', $SNAME, "\r\n";
print sprintf('Content-Length: %d', length($html)), "\r\n";
print 'Content-Type: text/html', "\r\n";
print "\r\n";
print $html;


sub enc_to_img {
    no bignum;

    my $encref = shift;

    my @enc = @{$encref};

    my $len = scalar(@enc);

    my $cols = int(sqrt($len));
    if ($cols > $MAX_COLS->numify()) {
        $cols = $MAX_COLS->numify();
    }
    my $rows = int($len / $cols);
    if ($cols * $rows < $len) {
        $rows++;
    }

    my $gridw = int((2 * ($R->numify() * $W->numify())) + $W->numify()); # each grid cell width

    my $expected_w = $gridw * $cols;
    my $scale_f = ($OUT_W->numify() * 1.0) / ($expected_w * 1.0);

    my $scaled_gridw = int($gridw * $scale_f);
    my $scaled_w = int($W->numify() * $scale_f);
    my $scaled_h = int($H->numify() * $scale_f);
    my $scaled_r = int($R->numify() * $W->numify() * $scale_f);

    #warn sprintf('gridw: %d, cols: %d, scale_f: %0.3f, scaled_gridw: %d, scaled_w: %d, scaled_h: %d, scaled_r: %d', $gridw, $cols, $scale_f, $scaled_gridw, $scaled_w, $scaled_h, $scaled_r), "\n";

    my $rn = 0; # ring number

    my $cmd = sprintf('convert -set colorspace Gray -size %dx%d xc:white', $cols * $scaled_gridw, $rows * $scaled_gridw);

    for(my $g = 0; $g < $len; $g++) {
        my @c = @{$enc[$g]};

        my $gx = $g % $cols;
        my $gy = int(($g - $gx) / $cols);

        my $num = scalar(@c);

        for(my $r = 0; $r < $num; $r++) {
            my ($rx, $ry) = (0, 0);

            if ($num > 1) {
                # Shrink R based on 6
                my $new_r = $scaled_r * sqrt(($num / 6.0));
                ($rx, $ry) = (int($new_r * sin(($r * 2.0 * pi) / $num)), -1 * int($new_r * cos(($r * 2.0 * pi) / $num)));
            }
            #warn 'rx: ', $rx, ' ry: ', $ry, "\n";

            my $truex = int(((($gx + 0.5) * $scaled_gridw) + $rx) - ($scaled_w / 2.0));
            my $truey = int(((($gy + 0.5) * $scaled_gridw) + $ry) - ($scaled_h / 2.0));

            #warn 'truex: ', $truex, ' truey: ', $truey, "\n";

            my $rfile = sprintf('%s/ring_%02d.png', $IDIR, $rn % $rings);
            $rn++;

            my $rcmd = sprintf(' \\( %s -resize %dx%d\\! -background white -rotate %d -crop %dx%d -repage +%d+%d -compose Multiply \\)', $rfile, $scaled_w, $scaled_h, $c[$r], $scaled_w, $scaled_h, $truex, $truey);

            $cmd .= $rcmd;
        }
    }

    $cmd = $cmd . sprintf(' -layers flatten -alpha off -resize %dx%d\\! png:-', $OUT_W->numify(), int($OUT_W->numify() * (($rows * 1.0) / ($cols * 1.0))));

    #warn 'cmd: ', $cmd, "\n";

    my $ret = `$cmd`;
}


sub img_to_html {
    my $img = shift;

    my $html = sprintf('<div class="alien"><p><img src="data:image/jpeg;base64,%s" alt="" /></p></div>', "\n" . encode_base64($img));

    return $html;
}


sub num_to_enc {
    my $n = shift;
    my $aclus = shift;
    my $aseq = shift;

    # Cluster list
    my @cl = split(//, $aclus);
    my $ci = 0;

    # Sequence list
    my @sl = split(//, $aseq);
    my $si = 0;

    my @enc = ();

    while ($n > 0) {
        #warn 'Working on n = ', $n, "\n";
        my @c = ();
        for (my $i = 0; $i < $cl[$ci]; $i++) {
            push @c, $sl[$si];

            $si++;
            $si %= scalar(@sl);
        }

        my $m = lcm_list(@c);

        my $d = $n % $m;

        my @r = ();
        foreach my $w (@c) {
            my $k = $d % $w;

            push @r, int((360 * $k) / $w);
        }

        push @enc, [@r];

        $ci++;
        $ci %= scalar(@cl);

        $n = ($n - $d) / $m;
    }

    return \@enc;
}


sub bin_to_num {
    my $b = shift;

    return -1 unless ($b =~ m/^[01]+$/);

    my @bits = split(//, $b);
    unshift @bits, '1';

    my $n = 0;

    foreach my $l (@bits) {
        my $c = ord($l) - ord('0');

        $n = ($n * 2) + $c;
    }

    #print 'word ', $w, ' -> ', $n, "\n";

    return $n;
}


# d = a * x + b * y  where d is the GCD
sub euclid_gcd {
    my $a = shift;
    my $b = shift;

    if ($b == 0) {
        return ($a, 1, 0);
    }

    my ($d2, $x2, $y2) = euclid_gcd($b, $a % $b);

    my ($d, $x, $y) = ($d2, $y2, $x2 - (int($a / $b) * $y2));

    return ($d, $x, $y);
}


sub lcm {
    my $a = shift;
    my $b = shift;

    my ($d, $x, $y) = euclid_gcd($a, $b);

    return ($a * $b) / $d;
}


sub lcm_list {
    my @l = @_;

    my $len = scalar @l;

    if ($len == 0) {
        die 'Can not LCD an empty list!', "\n";
    } elsif ($len == 1) {
        return $l[0];
    }

    my $m = $l[0];
    for (my $i = 1; $i < $len; $i++) {
        $m = lcm($m, $l[$i])
    }

    return $m;
}



# $ convert -size 2000x2000 xc:white \( ring_01.png -background white -rotate 30 -crop 950x950 -repage +500+500 -compose Multiply \) \( ring_01.png -background white -rotate 90 -crop 950x950 -repage +550+550 -compose Multiply \) -layers flatten -resize 500x500\! png:test.png
