#
Summary

Flight is a binary exploitation challenge which acts as a navigation computer for a spaceship. The final prompt is vulnerable to a buffer overflow. Since the binary has DEP enabled, ROP must be used to make a call to `system` and get a reverse shell. Since ASLR isn't enabled, the address for the `destination` variable, which is controlled by the user, can be determined before runtime and will act as the parameter to `system`. 

