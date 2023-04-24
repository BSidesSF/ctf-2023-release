#!/bin/bash

socat TCP-LISTEN:62144,reuseaddr,fork EXEC:./codebook.pl
