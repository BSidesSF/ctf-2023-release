require './libronsolve.rb'

system("javac -cp '.:../distfiles/FlatWhite.jar' Solve.java") || exit
flag = `java -cp '.:../distfiles/FlatWhite.jar' Solve`.chomp

check_flag(flag, terminate: true)
