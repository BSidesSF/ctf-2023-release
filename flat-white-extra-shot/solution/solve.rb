require './libronsolve.rb'

system("javac -cp '.:../distfiles/FlatWhiteExtraShot.jar' Solve.java") || exit
flag = `java -cp '.:../distfiles/FlatWhiteExtraShot.jar' Solve`.chomp

check_flag(flag, terminate: true)
