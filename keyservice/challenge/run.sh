#!/bin/bash

socat TCP4-LISTEN:1031,reuseaddr,fork EXEC:./keyservice,pty,stderr,setsid,sane
