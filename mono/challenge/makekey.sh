#!/bin/bash

perl -e 'for(my $i = ord("a"); $i <= ord("z"); $i++) { print chr($i), "\n";}' | sort -R | xargs echo -n | sed -r 's/ //g' | perl -ne 'print $_, uc($_), "\n";'

