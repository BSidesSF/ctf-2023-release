#!/usr/bin/perl

use strict;
use warnings;

use MIME::Base64;
use Digest::SHA qw(sha256_hex);

use bignum upgrade => undef; # Keep things as integers

my $W = 3562;
my $H = 2372;

my $W_out = 1024;
my $H_out = 680;

my $xo = 872; # x offset into mask

my $TW = 299; # Tile width
my $TH = 131; # Tile height

my $IDIR = '.';

my $SNAME = 'symmetric\'s truly horrible perl hack to approximate a webserver';

my $HTML_START = << 'ENDHTMLSTART';
<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="content-type" content="text/html; charset=UTF-8">
    <title>codebook</title>

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
      .books {
          float:center;
          text-align:center;
      }
      .form {
          color: white;
          font-size: 16pt;
      }
      .back {
          background-image: repeating-linear-gradient(rgba(52, 46, 34, 1), rgba(48, 36, 24), rgba(52, 46, 34, 1));
      }

      img {
          border: 5px solid #14150f;
      }
    </style>

  </head>
  <body class="back">
    <div class="form">
         <form method="get" action="/">
               <p>Key:<br /><input type="text" name="key" value="%s" size="87" maxlength="87" /></p>
               <p>Message:<br /><input type="text" name="msg" value="%s" size="32" maxlength="32" /></p>
               <p><input type="submit" name="click" value="Encrypt" /></p>
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


#make_key();

# http://127.0.0.1:1234/?key=3OKw89wWk3dwySCYOibzU6cVKHm6Ak90j7lEktPL7gTBaDd5CMjgnbzNGkbsErnQ71LE3SJ5vRV2V7eYYUYdz&msg=CTF{my_library_stores_516_bits}

alarm 5;

my $lc = 0;
my $got_get = 0;
my $error = 0;
my $hc = 0;
my $key = '';
my $msg = '';
while (<STDIN>) {
    $lc++;

    my $line = $_;

    if ($lc == 1) {
        #warn 'line 1', "\n";
        #warn $line, "\n";
        #if ($line =~ m/^GET\s\/\?key=([0-9A-Za-z]{1,87})&msg=(.{0,32})\sHTTP\/1\.[01]\s+$/) {
        if ($line =~ m/^GET\s\/\?.*?\sHTTP\/1\.[01]\s+$/) {
            $got_get = 1;
            if ($line =~ m/key=([0-9A-Za-z]{1,87})(?=(?:&|\sHTTP))/) {
                $key = $1
            } else {
                $got_get = 0;
            }

            if ($line =~ m/msg=([^&]{0,32})(?=(?:&|\sHTTP))/) {
                $msg = $1
            } else {
                $got_get = 0;
            }

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

my $keynum;
my @booklist = ();
if ($error == 0) {
    $keynum = key_to_number($key);

    if ($keynum != -1) {

        @booklist = number_to_base($keynum, 144);
    } else {
        $error = 1;
    }
}

# Pad out book list with 0s until it's 72 digits
for (my $i = scalar(@booklist); $i < 72; $i++) {
    unshift @booklist, 0;
}

my $keybin;
my $msg_comment;
if ($error == 0) {
    $keybin = pack("H*", sha256_hex(join('_', @booklist)));
    $msg_comment = '';
    my $mlen = length($msg);
    if (($mlen < 1) || ($mlen > 32)) {
        $error = 1;
    } else {
        $msg_comment = unpack('H*', $msg ^ (substr($keybin, 0, $mlen)));
    }
}


if ($error == 1) {
    print 'HTTP/1.0 303 See Other', "\r\n";
    print "Server: ", $SNAME, "\r\n";
    print 'Location: /?key=error&msg=abusewillnotbetolerated', "\r\n";
    print "\r\n";

    exit 0;
}


my $html = sprintf($HTML_START, $key, $msg);

$html .= "\n" . img_to_html(key_to_img(\@booklist, $msg_comment)) . "\n";


$html .= $HTML_END;

print 'HTTP/1.0 200 OK', "\r\n";
print 'Server: ', $SNAME, "\r\n";
print sprintf('Content-Length: %d', length($html)), "\r\n";
print 'Content-Type: text/html', "\r\n";
print "\r\n";
print $html;


sub key_to_img {
    my $bookref = shift;
    my $comment = shift;

    #warn 'comment: ', $comment, "\n";
    #warn 'books: ', join(', ', @{$bookref}), "\n";

    my $cmd = sprintf('convert -size %dx%d xc:white', $W, $H);

    for (my $i = 0; $i < 72; $i++) {

        my $x = int($i % 6);
        my $y = int($i / 6);

        my $b = $bookref->[$i];

        my $flop = '';
        if ($b >= 72) {
            $flop = '-flop';
            $b -= 72;
        }

        my $bx = int($b % 6);
        my $by = int($b / 6);

        my $file = sprintf('library_books_grid-%02d-%02d.png', $bx, $by);

        my $tx = $xo + $x * $TW;
        my $ty = $y * $TH;

        my $tile = sprintf(' \\( %s/%s %s -repage +%d+%d \\)', $IDIR, $file, $flop, $tx, $ty);

        $cmd = $cmd . ' ' . $tile;
    }


    # Put the front mask on
    my $mask = sprintf(' \\( %s/%s -repage +%d+%d \\)', $IDIR, 'library_front_mask.png', 0, 0);
    $cmd = $cmd . ' ' . $mask;

    $cmd = $cmd . sprintf(' -layers flatten -resize %dx%d\\! -set comment "%s" jpeg:-', $W_out, $H_out, $comment);

    my $ret = `$cmd`;

    return $ret;
}


sub img_to_html {
    my $img = shift;

    my $html = sprintf('<div class="books"><p><img src="data:image/jpeg;base64,%s" alt="" /></p></div>', "\n" . encode_base64($img));

    return $html;
}


sub key_to_number {
    my $w = shift;

    return -1 unless ($w =~ m/^[a-zA-Z0-9]{1,87}$/);

    my @lets = split(//, $w);

    my $n = 0;

    foreach my $l (@lets) {
        my $c = 0;
        if ($l =~ m/^[0-9]$/) {
            $c = ord($l) - ord('0');
        } elsif ($l =~ m/^[A-Z]$/) {
            $c = (ord($l) - ord('A')) + 10;
        } elsif ($l =~ m/^[a-z]$/) {
            $c = (ord($l) - ord('a')) + 10 + 26;
        } else {
            return -1;
        }

        # base 62 baby!
        $n = ($n * 62) + $c;
    }

    #print 'word ', $w, ' -> ', $n, "\n";

    return $n;
}


sub number_to_base {
    my $n = shift;
    my $b = shift;

    return -1 if ($n < 0);

    my @digits = ();
    my $t = $n;
    do {
        #warn 't: ', $t, "\n";
        my $d = $t % $b;
        #warn 'd: ', $d, "\n";

        unshift(@digits, $d);

        $t = ($t - $d) / $b;

    } while ($t > 0);

    return @digits;
}


sub make_key {


#GP/PARI> n = fromdigits([0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71],b=144)
#%3 = 12343188344305677644192283397683745422241744023540505380711488139489752297356739390398726556711855725038931925919536025384907864090179561659435005983911
#GP/PARI> digits(n, 62)
#%4 = [3, 24, 20, 58, 8, 9, 58, 32, 46, 3, 39, 58, 60, 28, 12, 34, 24, 44, 37, 61, 30, 6, 38, 31, 20, 17, 48, 6, 10, 46, 9, 0, 45, 7, 47, 14, 46, 55, 25, 21, 7, 42, 29, 11, 36, 13, 39, 5, 12, 22, 45, 42, 49, 37, 61, 23, 16, 46, 37, 54, 14, 53, 49, 26, 7, 1, 21, 14, 3, 28, 19, 5, 57, 27, 31, 2, 31, 7, 40, 34, 34, 30, 34, 39, 61]

    my $n = 12343188344305677644192283397683745422241744023540505380711488139489752297356739390398726556711855725038931925919536025384907864090179561659435005983911;

    warn 'num: ', $n, "\n";
    my @d144 = number_to_base($n, 144);
    my @d = number_to_base($n, 62);

    warn 'digits 144: ', join(', ', @d144), "\n";
    warn 'digits 62: ', join(', ', @d), "\n";

    my $k = '';

    foreach my $l (@d) {
        if ($l < 10) {
            $k .= chr($l + ord('0'));
        } elsif ($l < 36) {
            $k .= chr(($l - 10) + ord('A'));
        } else {
            $k .= chr(($l - 36) + ord('a'));
        }
    }

    warn 'key: ', $k, "\n";
}
