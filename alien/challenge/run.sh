#!/bin/bash

socat TCP-LISTEN:5440,reuseaddr,fork EXEC:./alien.pl
