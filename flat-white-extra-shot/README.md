We use a semi-obfuscated Java string here. We can't possibly make it SECURE
secure, but we can make it more difficult than casual inspection at least.

To build the obfuscated string:

```
$ cd challenge/simple-string-obfuscator
$ javac SimpleStringObfuscator.java && java SimpleStringObfuscator 'CTF{stronger-java-everywhere}'
```

Then copy it into the `System.out.println()` on Flag.java.

To build the project:

```
$ cd challenge
$ ant
```
