# Summary

This challenge is a white box web application challenge consisting of a web application written using Flask with a Sqlite database. The attacker has to read through the source code to realize that the the data provided for the registration endpoint can also be sent using JSON. Then by reading the client side code, see that the JSON for the registration objects on the admin panel are being merged using an insecure function which is vulnerable to prototype pollution. This can be used to pollute the object and cause XSS. This can then be leveraged to steal the admin cookie and bypass authentication. 

