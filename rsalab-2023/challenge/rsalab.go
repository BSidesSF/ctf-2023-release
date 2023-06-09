package main

import (
	"bufio"
	//"bytes"
	cryptor "crypto/rand"
	"math/big"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"syscall"
)

var (
	flag_2022_first, flag_2022_second, flag_2023 string
	sample_code string
)


const challenge_count = 11

var challenge_status [challenge_count]bool

const prog_name = "RSA Lab 2023"

const help_text = `
Commands:
    help                 // Prints this help
    help first           // Read this first
    help 2022            // Help on last year's challenges
    help 2023            // See what's new for 2023
    help conventions     // Describe the conventions used by this challenge lab
    help rsa             // Explanation of RSA
    help mod             // Explanation of modular arithmetic
    help code            // Dump sample code to help with some challenges
    help challenge <n>   // Get help on a specific challenge

    status               // Print the challenge statuses
    challenge <n>        // Complete the nth challenge

    getflag 2022-first   // Flag for completing first half of 2022 challenges
    getflag 2022-second  // Flag for completing second half of 2022 challenges
    getflag 2023         // Flag for completing the 2023 challenges

    exit                 // Exit the lab
    ^d                   // Same as exit but 100% more Unix-points
`

const help_conventions_text = `
RSA operates on numbers, not "messages" or "data". To encrypt or
decrypt data with RSA you must convert that data into a number. The
details of how this conversion process works is irrelevant to RSA.  To
simplify matters, this tool simply treats everything as a number.
Messages are numbers, ciphertext are numbers, etc. You don't have to
think about the meaning of these numbers in terms of data at all.

This tool uses decimal integer numbers exclusively. Every answer is
just a single number which should be entered as ASCII characters
representing the decimal digits of the number. Don't enter spaces,
commas, decimal points, or any other non-digit ASCII character.
`

const help_first_text = `
Welcome to the RSA Lab! This is not your "typical" CTF challenge! This
challenge won't play any games with you and especially won't try to
confuse you. This RSA "laboratory" is meant to make RSA-based
challenges as approachable and "easy" as possible.  The goal here is
that anyone with a willingness to learn and isn't afraid to use a
calculator (probably in the form of Python code) will solve all of the
challenges here.

If you've played any CTFs before you know that RSA-based challenges are
extremely popular. I (symmetric) have personally made at least 20 RSA
challenges for CTFs (sorry!). If you've played any of the past 5
years of BSidesSF CTFs you've probably seen one of my RSA challenges.

What I've learned from many previous RSA challenges is that >95% of
players look at them and go "hell no, this is an RSA challenge, I'm
going to work on something else". Of the remaining 5% of players that
do work on the RSA challenge, many seem to be graduate students whose
field of research involves cryptography. As someone with no formal
academic training in cryptography I often think "this is a SUPER HARD
challenge, it will take a skilled player at least 8 hours to solve"
and then 1-3 players solve it in 30 minutes like it was nothing, while
the remaining 99% of players never solve it.

This RSA lab is different. Every single challenge here is designed to
be approachable. Ample help text has been written about RSA and each
challenge. Sample code is given too! At no point in any of these
challenges are you supposed to wonder "what do I do?".

But that also doesn't mean these challenges are easy! The mathematics
(specifically number theory) behind RSA is complicated. The tools you
will use to solve these challenges will likely be new to you. You will
likely get pushed slightly out of your usual "comfort zone" in the
process of solving these.

The hope is that you will have fun, earn three flags, and learn
something in the process!
`


const help_2022_text = `
This "RSA Laboratory" set of challenges were first introduced for the
2022 BSides SF CTF. When I was first making the set I wanted to add
more challenges but I feared overwhelming players with "too much RSA"
(if there is such a thing!), and also, I kinda rand out of time.

But, after the set was well received, I thought it would be good to
expand it for 2023.

Challenges 1 through 8 were released in 2022 and challenges 9 through
11 are new for this year.

This year the 2022 challenges will get you two flags and the 2023
challenges will get you a third flag.

If you solved all of the challenges last year you will have no trouble
solving the three new ones this year too. If you relied on existing
tools and algorithms last year, consider building your own set of
tools this year. You can start with the Extended Euclidean Algorithm
for GCD & Chinese Remainder Theorem.

If you want to see the source code for 2022 it is available at
https://github.com/BSidesSF/ctf-2022-release/blob/main/rsalab/challenge/rsalab.go

We hope you have fun again this year and learn something new!
`

const help_2023_text = `
There are three new challenges this year (9, 10, and 11).

Challenge 9 is especially easy. Challenge 10 is
intermediate. Challenge 11 is probably the hardest challenge in the
set.

Feel free to rely on the help text as much as you'd like, that's why
it's there!
`

const help_rsa_text = `
Fully understanding RSA is a daunting task. RSA relies on somewhat
complicated number theory and many articles (like Wikipedia) that
explain RSA lean heavily on mathematical jargon that will likely
confuse you if you don't already have a number theory background.

Instead of trying to explain "why" RSA works, this help will focus on
the basic "how" RSA works. Fortunately, how it works is based on
rather simpler mathematics. Later, if you're curious, you can delve
into the background details of why RSA works.

RSA relies on three numbers: n, e, and d:

n is a big number called the modulus and it's public knowledge
e is the encrypting exponent and is also public
d is the decrypting exponent and it is private (secret)

Together n and e form the "public key" while d is the "private key".

The message to be encrypted is usually called m and an encrypted
message (the ciphertext) is called c.

Following these conventions encryption is:
c = m ^ e mod n

And decryption is the same operation using d instead:
m = c ^ d mod n

The key to making this all work is in how n was chosen and how d was
derived from e. n is made by picking two large primes, p and q, and
multiplying them together: n = p * q

e is generally made by choosing a number like 65537

d is found using e, p, and q. Importantly:

d * e = 1 mod ((p - 1) * (q - 1))

In other words, d and e are multiplicative inverses mod ((p - 1) * (q
- 1)). Don't worry if you don't understand this point just yet.

At first you will just "follow the rules" for picking p and q to make
n, choosing e, and finding d.

If you are curious about why (p - 1) * (q - 1) is used for the modulus
instead of just n, see
https://crypto.stackexchange.com/questions/1789/why-is-rsa-encryption-key-based-on-modulo-varphin-rather-than-modulo-n
for an excellent explanation.
`

const help_mod_text = `
The real trick to RSA is that everything happens modulo a number,
called "n" (also called the modulus). Modular arithmetic is often
called "clock arithmetic" but I don't think that's a super helpful
conceptual idea.

Modular arithmetic is division with a remainder. 10 mod 7 is the
remainder of 10 when divided by 7 which is 3.

I like to think of modular arithmetic as a loop where when you get to
the end, you just start over at the beginning.

For example, if you count to 10 without any modular arithmetic you
get: 0 1 2 3 4 5 6 7 8 9 10

If you count to 10 modulo 3 you get:
0 1 2 0 1 2 0 1 2 0 1

If you count to 10 modulo 7 you get:
0 1 2 3 4 5 6 0 1 2 3

The pattern to counting modulo a number should be pretty obvious. RSA
relies on the fact that numbers loop like this in modular arithmetic.
For example, if we add 1 + 2 = 3 mod 7 we get 3. But we also get the
same result if we add 4 + 6 = 3 mod 7. When working modulo 7 we can't
tell the difference in the result between 1 + 2 and 4 + 6 because they
both result in 3 (mod 7). RSA doesn't use addition like in this
example but the concept is the same. The fact that you can't tell the
difference between 1 + 2 and 4 + 6 means there is some ambiguity about
which numbers were used in the addition. RSA relies on the same sort
of ambiguity where you don't know what number was started with after
the RSA encryption operation.

In most of these challenges you will also have to compute the modular
multiplicative inverse of a number mod another number. For example,
the inverse of 2 mod 7 is 4 mod 7 because 2 * 4 = 1 mod 7. The
Wikipedia article explains more but doesn't provide a very good
description of how to find inverses.
https://en.wikipedia.org/wiki/Modular_multiplicative_inverse

See https://en.wikipedia.org/wiki/Extended_Euclidean_algorithm for a
more practical discussion plus an algorithm.
`

const challenge_1_name = "Getting Started: RSA Encryption"

const challenge_1_text = `
Before your start on attacking RSA, it's important to first understand
basic RSA operations like encryption and decryption. For this
challenge, we have generated a public key. Your job is to encrypt the
provided message with that public key to get the ciphertext.

RSA encryption uses the encrypting exponent, e. Decryption is the same
mathematical operation except that it uses the decrypting exponent, d.
If you can encrypt with the provided e then you could decrypt (if you
had d) just as easily.
`

const challenge_1_help = `
In this challenge you are given m which is the number 1234567890 and
you must raise it to the 65537 power mod the given n.

Suppose the n given to you is 89646143396657320858388540992681730174474568706914185426625368068570563407611

Using the python pow(m, e, n) function you could do this in one step:

symmetric@lambda ~ $ python -q
>>> pow(1234567890, 65537, 89646143396657320858388540992681730174474568706914185426625368068570563407611)
57567941894500688313929604728515154113526234207517454975448093522510002278562
>>>

Personally my favorite tool for the job is "gp" (the GP/PARI
calculator). I use the following function for modular exponentiation:

modexp(a, b, n) = { \
    my(d, bin); \
    d = Mod(1, n); \
    bin = binary(b); \
    for (i = 1, length(bin), \
         d = sqr(d); \
         if (bin[i] == 1, \
             d = d*a; \
         ); \
    ); \
    return(d);
}


symmetric@lambda ~ $ gp -q
GP/PARI> modexp(a, b, n) = { \
    my(d, bin); \
    d = Mod(1, n); \
    bin = binary(b); \
    for (i = 1, length(bin), \
         d = sqr(d); \
         if (bin[i] == 1, \
             d = d*a; \
         ); \
    ); \
    return(d);
}
GP/PARI> lift(modexp(1234567890, 65537, 89646143396657320858388540992681730174474568706914185426625368068570563407611))
57567941894500688313929604728515154113526234207517454975448093522510002278562
GP/PARI>

I like GP/PARI because it comes with ready-made functions for finding
primes and other useful mathematical tools.
`

const challenge_2_name = "Modular Inverse"

const challenge_2_text = `
One of the most intimidating things about RSA when you first read
about it is the notion of a "modular inverse". In our everyday lives
we're pretty familiar with the idea that (1/3) * 3 = 1 or that (5/7) *
(7/5) = 1.  That is, 1/3 is the inverse of 3 and 5/7 is the inverse of
7/5.  What happens though when the numbers "wrap around" because of a
modulus? As one example, you can verify easily that the inverse of 2
mod 5 is 3 mod 5 because 2 * 3 = 6 and 6 mod 5 = 1. The idea that 2
and 3 are inverses mod 5 is not at all obvious. Or, even if mod 5 is
obvious, in the general case where the modulus is large inverses are
not obvious.

In this challenge you are given e and n (as well as the factors of n,
p and q)and you need to find the modular inverse of e mod
phi(n). Recall that when n has two factors phi(n) is (p - 1) * (q - 1).
`

const challenge_2_help = `
This challenge almost generates an entire RSA key for you but leaves
the computation of d to you.

The easiest thing to do here is to use a tool like python or GP/PARI
to compute the inverse for you without fully understanding the
algorithm.  Later you can check out
https://en.wikipedia.org/wiki/Modular_multiplicative_inverse and
https://en.wikipedia.org/wiki/Extended_Euclidean_algorithm to see more
details.

Suppose this is the challenge:

== Generated Public Key ==
Public Modulus (n): 106774734048728380481835701597527292467011403920231274680839097590187865977691
Encrypting Exponent (e): 65537
== Secret Primes ==
Prime p: 333716405919027383574740039484765824347
Prime q: 319956502452073320086202721924536125953

So we need to find the inverse of e mod (p - 1)*(q - 1).  We can do
that in GP/PARI like so:

symmetric@lambda ~ $ gp -q
GP/PARI> e = 65537
65537
GP/PARI> p = 333716405919027383574740039484765824347
333716405919027383574740039484765824347
GP/PARI> q = 319956502452073320086202721924536125953
319956502452073320086202721924536125953
GP/PARI> d = lift(Mod(1/e, (p - 1)*(q - 1)))
105764612512646295615291030114388501839244164732852082244784995911009323602945
GP/PARI>


Instead using python (requires python >= 3.8):

symmetric@lambda ~ $ python -q
>>> p = 333716405919027383574740039484765824347
>>> q = 319956502452073320086202721924536125953
>>> e = 65537
>>> d = pow(e, -1, (p - 1)*(q - 1))
>>> d
105764612512646295615291030114388501839244164732852082244784995911009323602945
>>>
`

const challenge_3_name = "Choose-Your-Own-Key!"

const challenge_3_text = `
RSA is an "asymmetric" encryption algorithm because the key used to
decrypt (d) is not the same key that is used to encrypt (e). This
challenge will let you experience that first-hand. You provide your
public key (n) and (e) and this challenge will encrypt a secret
message (m) and provide you the ciphertext (c). Then using your
decrypting exponent (d) you can decrypt the ciphertext back to the
secret message.

You are highly encouraged to generate your own key here instead of
finding RSA keys somewhere.

To pass this challenge the n value you provide must be at least 512
bits, not be prime and the e value you choose must be at least
100. 65537 is a good choice for e.
`

const challenge_3_help = `
Here you need to generate your own RSA key. The first hurdle here is
that you need to pick two large primes. GP/PARI makes this easy:

GP/PARI> p = randomprime([2^256, 2^257])
140196833953916848276633857101089841829366176676275370027206281819893071421841
GP/PARI> q = randomprime([2^256, 2^257])
150391937610229902223804853235485922441763970021836916141449872542986006351799

Once you have p and q you can multiply them together to get n:

GP/PARI> n = p * q
21084473505149223825043760296349403402250329268423121096264322390309178612097102004892423269004740661799003897752631795192393692357560390726748390178241959

Now pick e (just choose 65537):

GP/PARI> e = 65537
65537

Now you need to find d which is the multiplicative inverse of e mod (p
- 1)*(q - 1). GP/PARI can do this for you quite easily:

GP/PARI> d = lift(Mod(1/e, (p - 1)*(q - 1)))
19753845273504546928933456306455127917078481886727315233100777562100399407378371267547139340609619053936970905555752219272636133969567174183728928194452833

Now if you enter n and e into challenge 2, you will be given an
encrypted message to decrypt:

"The ciphertext you must decrypt is 12934473471757421038542051093568093230211993619522507292825013065868993594496739155012700182447573492354927435828390239099132073979928048232811874908387314"

Use the 'modexp()' function provided by 'help code'.

GP/PARI> m = lift(modexp(12934473471757421038542051093568093230211993619522507292825013065868993594496739155012700182447573492354927435828390239099132073979928048232811874908387314, d, n))
165658259982062013333895374675803208468

So the message decrypted to the number
165658259982062013333895374675803208468


If you would rather use a tool like python, check out the "sympy"
package for generating the large primes:

symmetric@lambda ~ $ python -q
>>> import sympy
>>> sympy.randprime(pow(2, 256), pow(2, 257))
151242280369757164870463161646994705965267780422491921102587937827121481956083
`

const challenge_4_name = "Break RSA!"

const challenge_4_text = `
RSA is only secure when factoring n is hard. If you can factor n into
p and q then you can compute Euler's phi(n) (the totient function)
which, when n only has two prime factors, is just (p - 1) * (q -
1). With phi(n) you can use e to find d which gives you the private
key!

In this challenge, a secret message has been encrypted with a very
weak RSA key. The RSA key's n is approximately 128 bits meaning
(assuming they are the same size) p and q are about 64 bit each. It's
your job to decrypt the secret message by breaking the RSA key (factor
n!) and then determining d using n's prime factors. 128-bit RSA keys
are just big enough you can't use trivial methods like trial division
but are easily small enough that almost anything more sophisticated
will factor them in seconds.
`

const challenge_4_help = `
This challenge requires you to directly break the main security of
RSA: namely the difficulty of factoring large integers.

Fortunately there are many ready-made factoring tools and the n in
this challenge is small enough that even the slower ones will
(eventually) work.

GP/PARI is by far the fastest easy option:

symmetric@lambda ~ $ gp -q
GP/PARI> factor(286531825747973176248517811006933342689)

[16689809210752814749 1]

[17168070774791607061 1]


Python with sympy is another option:

symmetric@lambda ~ $ python -q
>>> from sympy.ntheory import factorint
>>> factorint(286531825747973176248517811006933342689)
{17168070774791607061: 1, 16689809210752814749: 1}
>>>


Even the 'factor' utility on Unix/Linux systems will work though it
will likely take several minutes:

symmetric@lambda ~ $ factor 286531825747973176248517811006933342689
286531825747973176248517811006933342689: 16689809210752814749 17168070774791607061


Once you've factored n this challenge is basically reduced to making
yourself a private key and decrypting a message which has been covered
by previous challenges.
`

const challenge_5_name = "Fermat's Big Factorization"

const challenge_5_text = `
RSA's security relies on factorization being a hard
problem. Unfortunately there are many different factorization
algorithms that are quite fast in certain special cases. This makes
generating secure RSA keys quite challenging because tons of gotchas
and pitfalls must be avoided. One pitfall is when p and q are too
close together because an algorithm known as "Fermat's Factorization"
can be used.

In this challenge, a 2048 bit RSA key has been generated using two
1024 bit primes, p and q. The flaw is that p and q share approximately
their first 500 bits. That may not sound that bad since they still are
different in their lower ~524 bits but you'll see that isn't nearly
enough.
`

const challenge_5_help = `
This challenge has the first big RSA key that you need to break (2048
bit!). If it weren't for the insecure generation of p and q this key
would be completely secure.

See a https://fermatattack.secvuln.info/ for a recent real-world
example of this factorization method in use to break real RSA keys.

As described in the challenge, p and q share a huge number of their
leading bits. Of course, you don't know what p and q are so you can't
immediately see that they share so many bits.  What you can see is that
n is quite close to a perfect square.

For example, 4*6=24 is close to a perfect square (5*5=25) so if we take the
sqrt(24) we see it's pretty close to an integer:

symmetric@lambda ~ $ gp -q
GP/PARI> sqrt(24)
4.8989794855663561963945681494117827839

Using an example n from this challenge we can see the same thing:

symmetric@lambda ~ $ gp -q
GP/PARI> \p 2000
   realprecision = 2003 significant digits (2000 digits displayed)
GP/PARI> n = 20986093983229740737616979428461869736283676305759855367699716088339023187789743483739494558830824994560327218561249595776728160897172742009162263718411941586881117780544720226522193311631305288562036408547316148076659060909259807652309741897999194417302182865826334863281014754437183397426529926052388339637137917283012009808149735558120129115437014563763216695205449595542432222092449072689016799618954293944871853454720277574323238743681158558116640225824873483365405669165053284447897847579778436093799145645491124859843295334896761938649291274274176500212451143898760024463777014542402997669209322759319062164329;
GP/PARI> printf("%.5f\n", frac(sqrt(n)))
0.93993

One way to factor a number n that is a product of only two primes is
to guess p and see if it divides n evenly. If it does then q is just
the result after dividing by your correctly guessed p. This is called
trial division.

Another way to factor a number is to guess the number perfectly
in between the two primes. Since the primes are both odd, there is
always an integer between them. Guessing the number between the two
primes is called Fermat's factorization method.

Simple algebra shows how this work. Suppose b is the number between p
and q and a is the difference between b and p and b and q.  Then p = b
- a and q = b + a. Writing this out:

(b - a)*(b + a) = n

Then if we expand this polynomial we get a difference of squares:
b^2 - a^2 = n

If we algebraically re-arrange we can see how to check if our b guess is
correct:
b^2 - n = a^2

So we guess b, square it, subtract n, and check if that is a perfect
square. If it is we know our b guess is correct and we can take the
sqrt() of the result to get a.

When p and q are close together then b is close to the sqrt(n) so we
start guessing b there.

Let's use this to factor 35. First take the square root:

GP/PARI> b = ceil(sqrt(35))
6

So we start at 6 for our guess for b.

GP/PARI> n = 35
35
GP/PARI> b^2 - n
1

So the result is 1 which is a perfect square (1^2 = 1). So a = 1.
This means p = 6 - 1 and q = 6 + 1. In other words p = 5 and q = 7.

(b - a)*(b + a) = (6 - 1)*(6 + 1) = 5 * 7 = 35

Of course in this example the first guess for b was already the
correct value between the two primes so only one step was needed.

For another example using the large n above we can search for b by
starting at the square root:

GP/PARI> n = 20986093983229740737616979428461869736283676305759855367699716088339023187789743483739494558830824994560327218561249595776728160897172742009162263718411941586881117780544720226522193311631305288562036408547316148076659060909259807652309741897999194417302182865826334863281014754437183397426529926052388339637137917283012009808149735558120129115437014563763216695205449595542432222092449072689016799618954293944871853454720277574323238743681158558116640225824873483365405669165053284447897847579778436093799145645491124859843295334896761938649291274274176500212451143898760024463777014542402997669209322759319062164329;
GP/PARI> b = sqrtint(n); for(o = 0, 2^24, if(issquare((b + o)^2 - n), printf("Found b: %d\n", b + o); break()))
Found b: 144865779200022738626242031463206827390887956308018455112709870392900115303749764988814526159022386623910922507157136008728468769820216765575463709414250928780172888938697852869744135866950810434674492200047370645887934235096204238857872700254246163314531125381608725094887473425223695300148858373657699138827

So we have b, we can find a with the sqrt():

GP/PARI> a = sqrtint(b^2 - n)
7550222542724957377723122082424095566049178436705109584428009034999218965249411881269563533611994078677302091491076838220898322668815807974197776623255181540

So p = b - a and q = b + a:

GP/PARI> p = b - a
144865779200022738626242031463206827390887956308018455112709870392900115303749764988814526159022386623910922507157136008728468769820216765575463709414243378557630163981320129747661711771384761256237787090462942636852935016130954826976603136720634169235853823290117648256666575102554879492174660597034443957287
GP/PARI> q = b + a
144865779200022738626242031463206827390887956308018455112709870392900115303749764988814526159022386623910922507157136008728468769820216765575463709414258479002715613896075575991826559962516859613111197309631798654922933454061453650739142263787858157393208427473099801933108371747892511108123056150280954320367

We can check that these multiply to n:
GP/PARI> p * q == n
1

Now that we have the factors for n deriving the rest of the key (d)
follows like previous challenges.
`

const challenge_6_name = "Weak Entropy: Weak Keys"

const challenge_6_text = `
It is critically important that the primes used in an RSA key have
never been used in any other RSA key. If two keys happen to share a
prime then a quick comparison using the GCD algorithm between the two
n values will find a common prime. This is a common flaw for keys
generated on devices with poor (or no) entropy on first boot.

This challenge generates two keys that share a prime. Use the GCD
between the two n to find the common prime.
`

const challenge_6_help = `
This challenge is another common problem for real-world RSA keys. See
https://factorable.net/weakkeys12.extended.pdf for an example of this
problem at internet-scale.

Solving this challenge is essentially trivial. You have two n values
that share a common prime. The GP/PARI gcd() function can do it:

GP/PARI> gcd(12345, 456789)
3

You might consider taking this time to implement the Extended
Euclidean algorithm for GCD because you will need it in a later
challenge. https://en.wikipedia.org/wiki/Extended_Euclidean_algorithm
`

const challenge_7_name = "Exponent Size Matters"

const challenge_7_text = `
The modular arithmetic aspect of RSA is one of the main security
features that prevent trivial decryption of messages. It's important
that the modulus n reduces the resulting value's size. For example, if
the encrypting exponent were 3 and the ciphertext were 125 then a
simple cube root would reveal that the encrypted message was
5. Because of this, RSA is generally considered insecure when used
"raw" like in this challenge. Messages are supposed to be padded in
size so that 5 becomes something like 100000000...000005. Exact
padding schemes vary and are supposed to have some randomness to
them. Also, the encrypting exponent needs to be big enough that
encryption operations produce results many times the size of n so that
the modulus by n reduces them greatly. Finally, it's important that
the same message is not encrypted with multiple keys. Using "Chinese
Remainder Theorem" (CRT) a message encrypted with multiple keys can be
reconstructed if the sum of the modulus size for all the key exceeds
the size of the encrypted result (pre-modulus).

In this challenge, the message getting encrypted is 64 bits and the
encrypting exponent is 101. This produces a result that is
approximately 101 * 64 = 6500 bits. The message is then encrypted
with four different 2048 bit RSA keys.  The four keys combined are
about 8200 bits which is bigger than the 6500 bit intermediate result
of the encryption operation.  Using CRT the 6500 bit encrypted message
can be recovered and then the 101st integer root can be taken to
recover the original 64 bit message.
`

const challenge_7_help = `
This challenge makes use of an important result in number theory
called the "Chinese Remainder Theorem", or CRT for short.

CRT says that if you have a number mod a and another number mod b,
then as long as a and b are co-prime (the gcd(a,b) = 1) then there is
a unique number mod a*b with these remainders mod a and b.

Here is a simple example with some small numbers:

GP/PARI> chinese(Mod(5, 7), Mod(4, 5))
Mod(19, 35)

So 19 is the unique number less than 35 that is 5 mod 7 and 4 mod 5.

You can chain multiple applications of CRT together:

GP/PARI> chinese(chinese(Mod(1, 2), Mod(2, 3)), Mod(3, 5))
Mod(23, 30)

So 23 is the unique number less than 30 that is:
1 mod 2
2 mod 3
3 mod 5

Also note that 30 = 2 * 3 * 5

The sympy package for python also provides a CRT function:
symmetric@lambda ~ $ python -q
>>> from sympy.ntheory.modular import crt
>>> crt([2, 3, 5], [1, 2, 3])
(mpz(23), 30)

In this challenge you have a secret message that has been raised to
the 101st power mod four different n values. Using CRT you can combine
these four remainders mod the four different n values to get one huge
6500 bit number.  Once you've recovered m^101 you can take the 101st
root to recover the m value.

GP/PARI> secret = 3^101
1546132562196033993109383389296863818106322566003
GP/PARI> sqrtnint(secret, 101)
3
`

const challenge_8_name = "Signing Fault"

const challenge_8_text = `
One of the slowest aspects of RSA encryption and decryption operations
is the need to perform modular arithmetic mod n. The other operations
that are slow are the ones that use the decrypting exponent, d, since
d tends to be quite large like n.

One common optimization is to implement RSA operations mod p and mod q
and then use the Chinese Remainder Theorem to assemble the smaller
results into the final result mod n. This is called RSA-CRT.

Although RSA-CRT can be substantially faster than traditional RSA, it
requires saving p and q as part of the private key and then the
intermediate steps to perform an RSA operation use p and q. If a
hardware fault (like a bit flip) occurs during an operation with p or
q there is risk that p or q could get exposed in the final calculation
result.

In this challenge, two signatures are provided. One signature is valid
and the other signature was produced using RSA-CRT with a fault,
resulting in an invalid signature.

Using the faulty signature you need to recover the private key.

See https://eprint.iacr.org/2002/073.pdf for a discussion of this
issue and section 2.2 for specifics.
`

const challenge_8_help = `
First, to understand RSA-CRT check out this excellent Stack Exchange
answer https://crypto.stackexchange.com/a/2580/12803

See also section 2.2 of https://eprint.iacr.org/2002/073.pdf for
details on the key recovery.

Now we will work through an example using:

GP/PARI> n = 28023908291085975953552389149323829894383418678106405289275941263238623334709318452337865579668543451476379820199515189612119588195421403398769707784607777149665756091789789390029156713534780638684578717903107762236227711893650535436530914705040359219512820037684021347229164813034211160200006498054221559332775915692086376180415368912393767616873835239395015407988726810043557291200228661787519326071187281656845763099456827508174878937553677012116584219895568998037830743003339105116659823662867807579237567130650192383770115747597080242084867486677394590505844755274984337105763222837768103627722636676856338966527;
GP/PARI> e = 65537;
GP/PARI> s_good = 8303651532159008425670887194258057525871672890280171668608231415434231194930427903974399384085297662564475161842406801817018363017628204629672551440221515373741444639809144787882339744126115808860365604589535671658927684998514913642143537173358430254152131566481561995751432270905882123520637716031061088179122043644248548076039471388680788221737804317318692146472418306929597047709467041860845681866474756025023970300606844471605992776429713113941168833984636056425473616409989658398059634762080324775480856638193554262974552740247910895056554352269460208322350077305822857609687158247909462899941562902264622768375;
GP/PARI> s_bad = 19342994609554741458357554485074412839940628830015712044676375490208337290737758614751772962069056482374201452887284901395109068111452946639442280485996191851548315361898630048778037038756181737817069557484457082571840813517834446797214305761023883333560726389891779506737035582113337755454608817377002214259426811126220352614879478197158047018166923832232331333164650787867428732408360747478190701124642956648633495209708388099489350449604762559153650064603711573002067699131047918386310021977150379212903050550731044084725868664359435672256306530831047560432842215535389972425114013895434526326885255334529433798753;

First we recover the message from the good and bad signatures:

GP/PARI> m_good = lift(modexp(s_good, e, n))
62843116602763141656753934427477147059453069497663403989920794632807573006045
GP/PARI> m_bad = lift(modexp(s_bad, e, n))
11953066544292806019422985031256986963753003305148978814441907824144779784763826841815212343329476016399742776203031255298726352817016200942560960119593620488240832521263625149179255110223105543403709922071219558791933210169224221825359203188072766623314429769088903806512159455821748613842672032996610329703635438041837482118436161021571974216653839734881587618336220709238586989282060816215100701644508376674285924507486299312308743901314964030872521924455603148588255050673742852507937884909858938778123782305430609202712317152561940103848443349014273899638816116997646640312454559035507074371028094187257023997382

Now the good and bad messages differ by some multiple of one of the
prime factors of n:

GP/PARI> p = gcd(lift(Mod(m_good - m_bad, n)), n)
163202712121213342427812078030957671957343024218566452589912945591480258723291426639763729218993387558307370318535298096656951022145332412322132921726877173014222351661344787388684233764822830120796318445520017205306815388956855551573365542158774783595808434012279328405210647997424987345093104899589085991221
GP/PARI> q = n / p
171712270751187993003029622900346317611385036123454388354322495693485592657914944168367912259966890844882679023068692235901154611889231470940033721180897361577924205675952019529154949602919845027807224227648439921916358834135637733793660259992352199368448928406658057248380706107625984663255813928340601884387
GP/PARI> p * q == n

Now that we've recovered p and q we can generate the rest of the
private key.
`

const challenge_9_name = "Backwards Signing"

const challenge_9_text = `
RSA has a surprising property which is that you can find the
corresponding message for a signature by using public information.
This trick is sometimes called "backwards signing".

In RSA encryption and decryption are inverses of each other and
signing with RSA is "just" decryption (since decryption uses the
private key exponent).

That is, to sign a hash, h: s = h^d mod n
To verify the signature: h = s^e mod n

For backwards signing, you choose what you want the signature to be,
and then you encrypt it to find the corresponding message.

This feature of RSA is generally considered useless since you can't
sign a message of your choosing but instead find the message for a
corresponding signature of your choosing. This does mean though that a
user can't be trusted to be a key's owner (know the private key) just
because they can produce message, signature pairs.

Real RSA uses special padding and hashing for signatures to futher
protect against this attack.
`

const challenge_9_help = `
Using the provided n and e, choose any signature greater than 1.

Here using 2 for the signature:

GP/PARI> n = 27085323901953605912078181385460157493391920004099213242921637670412321516615285757305255991804124134025443964077465032773673285241239690754922635307494350431103514312259160916685969632448747868632058185207093085419857417320348344778819828276433751179288659067562873216605853689408410477774635640286817694838930385503116020831244467328142436933821343178730282206656367671229980087425266814567040822115341121422845740117280021795792596495669796092479389270984129618512238564794491641382803168766406233690077616443014599404998083176787904405258923824191452960506742771837933954725458040928123891852886121039576642828237

GP/PARI> e = 65537

GP/PARI> lift(modexp(2, e, n))
19551246001298344379187975772042513747294446992123450142376328004150394657607256423508373795964044740784029110187400958798462037416575655982046840200041858999864969147780750507828176890298557039732281720109747037254257822738178271258010738398768967547806245144085258289666502273237548484306633613654709853223526892292968825296967182025314447421052021463569350632399979200112788163239985891749601383214430676579381154008508477279881238450729848815312004612511247357206251486437803366752546643725590118117055122939904413888826956967909888151652743619990300832036818193901466992614106836348746431601268729447897774082465

So 2 is the signature for the long message starting with 195512...
`

const challenge_10_name = "Ciphertext Malleability - Signature Forgery"

const challenge_10_text = `

Many encryption schemes has a property knows an "malleability" where
mathematical operations on encrypted messages have a predictable
effect on the corresponding decrypted message.

For example, with XOR if you flip a bit in the ciphertext that
corresponds to the same bit being flipped in the plaintext.

For RSA the malleable property is multiplication.  That is, for two
messages, m1 and m2:

enc(m1, e, n) * enc(m2, e, n) == enc(m1 * m2, e, n)

Stated more concretely, the encryption (or signature) of 15 is the
same as the encryption (or sig) of 3 times the encryption (or sig) of
5.

In this challenge you are asked to generate a signature for a 256 bit
hash. You are given the ability to sign smaller messages. Use this
ability to construct a signature from the smaller factor's signatures.
`

const challenge_10_help = `

To solve this challenge you will need to facter the hash. The hash has
been constructed to make it especially easy to factor.

Once you have the factors, generate some signatures for them and then
mulitply the signatures together.

For example:

Hash to be signed (h): 7055649898239640755522448832898060754983923108720222515966576322755900417420

GP/PARI> factor(7055649898239640755522448832898060754983923108720222515966576322755900417420)

[             2 2]
[             5 1]
[             7 2]
[            13 1]
[            17 1]
[            31 2]
[          2621 1]
[          6653 1]
[         43399 1]
[        406699 1]
[       6024671 1]
[       8343029 1]
[  356214523661 1]
[  463763853197 1]
[13264543465981 1]

So you can sign 13264543465981, (463763853197 * 356214523661),
(8343029 * 6024671), and the remaining factors.

Enter a small 'hash' to be signed (limit 128 bits) (enter 0 to end signing loop): 13264543465981

Signature: 5222758239776843700960041289547027963971098727472131137512627384924651777556259468347807862526965315939551080344516638166027044521894198234048478996120960545126715891516838237508638483899212298051016254200635513678370157525968976935074654196030535655104800394049695270898270156648446994356094981113194162222803031057280312392474032292442415643802470028644329979037129174915787741453074386885183634087370114678505514416397834734356801658742563463737417636690920684810636154417305651776788647141768372373631499861533966728868861637927254130429619460535097045841640710249693059275358092653487334932167410069513816571406

Enter a small 'hash' to be signed (limit 128 bits) (enter 0 to end signing loop): 165199420057759286994217

Signature: 20891033479375525501083087029899855320692541301215765217516209112856121112074325859995222944997361160625807782571776903433648172980456161543825486255128750650067458050313507052324861594336035121230539134602323161819476810090636714781423497135452792779739956172811354463987877856024879947028520854327069449278450542166817304757530365444947277370920865651321574836986543856476688133796261201065816963696459567237132898899305054721617567272466313641918599567242293186107816867689386587726242774486480103991638497944330468935753279621785488626441200266876654889698545087027299397629339964183413137037014976072895662904926

Enter a small 'hash' to be signed (limit 128 bits) (enter 0 to end signing loop): 50264004868459

Signature: 17632493192155232693542939635311773079254030015734255574058552669697134524247898900015086507806155372279105033484870563184118906906818268645653808675631742095055445889513554922187687360612781154663533103435120909226302828458192629654160011716239940090469944580341842500818161526001979272156337915469324291095647991031517347055164583878487811698490422284512300724542947406476792072738844202469952349562413083928096635863879227783264934740062106267620296873867268410793594931763698083086578761165762719206153602338325283245516307118221825641558893378356640544404011288244269657349881722921127364169860536167428268443371

Enter a small 'hash' to be signed (limit 128 bits) (enter 0 to end signing loop): 64058845687999447271289940

Signature: 21174318812944618311453588119343120586334772049686330522339085731842057819430839834740986808155698920881057807702915732972752474998680103135983805230805381681547055273751051375173871577064153948013271125304856297446306758653795732703713844703197151299471692910286102579739020743142991591426487820115006946223110596437211939719780690444167260914676462563531503772168338689550802221626656112436270338124982279385843753461555681821809585149557914747194615199165553274998000268725232495352051565472629594982798381741802362904118827070251676669150708840814854886942505787626434600300473824678956335026195241238580543556096

Now we have 4 signatures. Multiply them all together (mod n) to get the final signature.

GP/PARI> lift(Mod(s1 * s2 * s3 * s4, n))
8528018079650986899429445668265573155770049829973260696728028915644938839027672282112899521227044208164370637501411210822088851426216517624450273593756213441801888822991881579800763473081504708544375274291888788999604684173543961140841319427020122556545515809450402401925904712513327644789823026471854866436482269033591102021569497968416639029191723471163357554812342372340691056339982321850516004292100548086639855607779368517815553438907173288780871610391134637394914233687579028725372683740222720536930214391636810403525679369536251296110798530224344480143729004592482748421491602566835028244479985630208801249370

`

const challenge_11_name = "Pollard's p - 1 factoring algorithm"

const challenge_11_text = `
In this challenge a key has been specially generated to be weak to
Pollard's p - 1 factorization algorithm.

Pollard's algorithm is efficient when one of its primes, p, is of a
special form. That is, that p - 1 only has small factors, not big
ones.  One small example of this is the prime 17. Since 17 - 1 is 16
and 16 is a power of 2, 16 only has small factors.

Using Pollard's p - 1 algorithm factor n to determine the private key
and decrypt the provided encrypted message.
`

const challenge_11_help = `
Pollard's p - 1 algorithm is perhaps the simplest specialized
factoring algorithm. That does not mean it's easy to understand at
first glance.

See https://en.wikipedia.org/wiki/Pollard%27s_p_%E2%88%92_1_algorithm

One key observation that makes the p - 1 algorithm work is that
working mod n is the same as working modulo the factors of n. For
example, working mod 6 is the same as working mod 2 and mod 3 at the
same time:

Starting mod 6:
5 + 4 mod 6 = 3 mod 6

Now mod 2:
5 mod 2 + 4 mod 2 = 1 + 0 = 1 mod 2

And mod 3:
5 mod 3 + 4 mod 3 = 2 + 1 mod 3 = 0 mod 3

Notice than 3 mod 3 is 0 and 3 mod 2 is 1, just as we got above.

Using Chinese Remainder Theorem (CRT) we can reconstruct 3 mod 6:

GP/PARI> chinese(Mod(1, 2), Mod(0, 3))
Mod(3, 6)


The other key observation is Fermat's Little Theorem which states:

a ^ (p - 1) = 1 mod p

This works for any a coprime to p, and any prime, p.

Notice here the p - 1 in the exponent. If we take any multiple of p -
1 we'll just be raising 1 to higher powers, which is still just 1.

Since when working with mod n we're actually working mod p and mod q,
if we raise a number to a multiple of p - 1 then that operation will
be 1 mod p.

In this challenge, the largest prime factor of p - 1 is guaranteed to
be less than 65535.

Using a toy example of n = 17 * 19:

GP/PARI> gcd(3^(2^8 % n) - 1, n)
17

Here we chose a to be 3 and we used some multiple of p - 1 which in
this case is 2^8.  3^(2^8) is 1 mod p so subtracting 1 and finding the
GCD with n yields p.

In this small example the only factor of p - 1 was 2. In a real
example, you will have to use every prime less than the given bound
(65535).

Here is a full attack in GP/PARI:
GP/PARI> pollard_p1(n) = {my(a = Mod(2,n)); my(g = 1); my(p = 2); while(g == 1, a = a^(p^floor(log(n)/log(p))); g = gcd(lift(a) - 1, n); p = nextprime(p + 1);); print(g);}

Note that Pollard's p - 1 algorithm isn't considered relevant
anymore. The chance that one of the primes in 2048 bit RSA has only
small factors is so low it's not worth worrying about. Furthermore,
there are other factoring algorithms that at at least as efficient as
the p - 1 algorithm across a much broader range of numbers.
`


func main() {
	startup()

	input := bufio.NewScanner(os.Stdin)
	scanbuffer := make([]byte, 65536)
	input.Buffer(scanbuffer, 65536)

	fmt.Fprint(os.Stdout, "\nWelcome to the RSA Laboratory! We hope you learn and have fun!\n")
	fmt.Fprint(os.Stdout, "\nTry \"help\" for a list of commands\n")

	exit := false

	for !exit {
		fmt.Fprintf(os.Stdout, "\n%s> ", prog_name)
		ok := input.Scan()
		if !ok {
			fmt.Fprintln(os.Stdout, "")
			break
		}

		text := input.Text()

		if len(text) == 0 {
			continue
		}
		//fmt.Fprintf(os.Stdout, "Got command: %s\n", text)

		tokens := strings.Split(text, " ")

		switch tokens[0] {

		case "help":
			if len(tokens) > 1 {
				switch tokens[1] {

				case "first":
					fmt.Fprintf(os.Stdout, "%s", help_first_text)

				case "2022":
					fmt.Fprintf(os.Stdout, "%s", help_2022_text)

				case "2023":
					fmt.Fprintf(os.Stdout, "%s", help_2023_text)

				case "rsa":
					fmt.Fprintf(os.Stdout, "%s", help_rsa_text)

				case "mod":
					fmt.Fprintf(os.Stdout, "%s", help_mod_text)

				case "conventions":
					fmt.Fprintf(os.Stdout, "%s", help_conventions_text)

				case "code":
					print_sample_code()

				case "challenge":
					if len(tokens) <= 2 {
						fmt.Fprintf(os.Stdout, "challenge help requires a challenge number\n%s", help_text)
					} else {
						switch tokens[2] {

						case "1":
							fmt.Fprintf(os.Stdout, "%s", challenge_1_help)

						case "2":
							fmt.Fprintf(os.Stdout, "%s", challenge_2_help)

						case "3":
							fmt.Fprintf(os.Stdout, "%s", challenge_3_help)

						case "4":
							fmt.Fprintf(os.Stdout, "%s", challenge_4_help)

						case "5":
							fmt.Fprintf(os.Stdout, "%s", challenge_5_help)

						case "6":
							fmt.Fprintf(os.Stdout, "%s", challenge_6_help)

						case "7":
							fmt.Fprintf(os.Stdout, "%s", challenge_7_help)

						case "8":
							fmt.Fprintf(os.Stdout, "%s", challenge_8_help)

						case "9":
							fmt.Fprintf(os.Stdout, "%s", challenge_9_help)

						case "10":
							fmt.Fprintf(os.Stdout, "%s", challenge_10_help)

						case "11":
							fmt.Fprintf(os.Stdout, "%s", challenge_11_help)

						default:
							fmt.Fprintf(os.Stdout, "challenge help requires a challenge number\n%s", help_text)
						}
					}

				default:
					print_help()
				}

			} else {
				print_help()
			}

		case "challenge":
			if len(tokens) == 2 {
				switch tokens[1] {

				case "1": // =========================== 1 ============================

					fmt.Fprintf(os.Stdout, "Challenge 1:\n%s\n", challenge_1_text)

					fails := 0
				ch_1_retry_key:
					p, err := cryptor.Prime(cryptor.Reader, 128)

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate prime p\n")
						os.Exit(1)
					}

					q, err := cryptor.Prime(cryptor.Reader, 128)

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate prime q\n")
						os.Exit(1)
					}

					n := new(big.Int).Mul(p, q)

					pm1 := new(big.Int).Add(p, big.NewInt(-1))
					qm1 := new(big.Int).Add(q, big.NewInt(-1))

					etot := new(big.Int).Mul(pm1, qm1)

					e := big.NewInt(65537)
					d := new(big.Int).ModInverse(e, etot)

					if d == nil {
						if (fails > 5) {
							fmt.Fprintf(os.Stderr, "Error: unable to generate d! Probably (p - 1) or (q - 1) was a multiple of e\n")
							os.Exit(1)
						} else {
							fails++
							goto ch_1_retry_key // OMG a GOTO!!1!!
						}
					}
					m := big.NewInt(1234567890)
					c := new(big.Int).Exp(m, e, n)

					fmt.Fprintf(os.Stdout, "== Generated Public Key ==\n")
					fmt.Fprintf(os.Stdout, "Public Modulus (n): %s\n", n.Text(10))
					fmt.Fprintf(os.Stdout, "Encrypting Exponent (e): %s\n", e.Text(10))
					fmt.Fprintf(os.Stdout, "\n")
					fmt.Fprintf(os.Stdout, "Message to encrypt (m): %s\n", m.Text(10))

					fmt.Fprint(os.Stdout, "\nWhat is the ciphertext (c) for the message (m) using the public key? ")
					ok := input.Scan()
					if !ok {
						fmt.Fprintln(os.Stdout, "Error reading input!")
						break
					}

					atext, ok := new(big.Int).SetString(input.Text(), 10)

					if !ok || atext == nil {
						fmt.Fprintln(os.Stdout, "Error parsing answer!")
						break
					}

					if atext.Cmp(c) == 0 {
						fmt.Fprintf(os.Stdout, "Correct!")
						challenge_status[0] = true
					} else {
						fmt.Fprintf(os.Stdout, "Unfortunately that is not correct.\n")
						fmt.Fprintf(os.Stdout, "The correct answer was %s\n", c.Text(10))
						fmt.Fprintf(os.Stdout, "Which could be found with m^e mod n:\n")
						fmt.Fprintf(os.Stdout, "%s ^ %s mod %s\n = %s\n", m.Text(10), e.Text(10), n.Text(10), c.Text(10))
					}


				case "2": // =========================== 2 ============================

					fmt.Fprintf(os.Stdout, "Challenge 2:\n%s\n", challenge_2_text)

					fails := 0
				ch_2_retry_key:
					p, err := cryptor.Prime(cryptor.Reader, 128)

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate prime p\n")
						os.Exit(1)
					}

					q, err := cryptor.Prime(cryptor.Reader, 128)

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate prime q\n")
						os.Exit(1)
					}

					n := new(big.Int).Mul(p, q)

					pm1 := new(big.Int).Add(p, big.NewInt(-1))
					qm1 := new(big.Int).Add(q, big.NewInt(-1))

					etot := new(big.Int).Mul(pm1, qm1)

					e := big.NewInt(65537)
					d := new(big.Int).ModInverse(e, etot)

					if d == nil {
						if (fails > 5) {
							fmt.Fprintf(os.Stderr, "Error: unable to generate d! Probably (p - 1) or (q - 1) was a multiple of e\n")
							os.Exit(1)
						} else {
							fails++
							goto ch_2_retry_key
						}
					}

					m, err := cryptor.Int(cryptor.Reader, new(big.Int).Exp(big.NewInt(2), big.NewInt(96), nil))

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate m\n")
						os.Exit(1)
					}

					c := new(big.Int).Exp(m, e, n)

					fmt.Fprintf(os.Stdout, "== Generated Public Key ==\n")
					fmt.Fprintf(os.Stdout, "Public Modulus (n): %s\n", n.Text(10))
					fmt.Fprintf(os.Stdout, "Encrypting Exponent (e): %s\n", e.Text(10))
					fmt.Fprintf(os.Stdout, "== Secret Primes ==\n")
					fmt.Fprintf(os.Stdout, "Prime p: %s\n", p.Text(10))
					fmt.Fprintf(os.Stdout, "Prime q: %s\n", q.Text(10))
					fmt.Fprintf(os.Stdout, "\n")

					fmt.Fprint(os.Stdout, "\nWhat is secret decrypting exponent (d) that is the couterpart of e? ")
					ok := input.Scan()
					if !ok {
						fmt.Fprintln(os.Stdout, "Error reading input!")
						break
					}

					dtext, ok := new(big.Int).SetString(input.Text(), 10)

					if !ok || dtext == nil {
						fmt.Fprintln(os.Stdout, "Error parsing answer!")
						break
					}

					// Test if provided d can decrypt the ciphertext
					// This test is used to allow for alternative ds like
					// using the Carmichael totient
					mdec := new(big.Int).Exp(c, dtext, n)

					if mdec.Cmp(m) == 0 {
						fmt.Fprintf(os.Stdout, "Correct!")
						challenge_status[1] = true
					} else {
						fmt.Fprintf(os.Stdout, "Unfortunately that is not correct.\n")
						fmt.Fprintf(os.Stdout, "The correct answer was %s\n", d.Text(10))
						fmt.Fprintf(os.Stdout, "Which could be found by finding 1/e mod phi(n):\n")
						fmt.Fprintf(os.Stdout, "%s * %s mod %s = 1\n", e.Text(10), d.Text(10), etot.Text(10))
					}

				case "3": // =========================== 3 ============================

					fmt.Fprintf(os.Stdout, "Challenge 3:\n%s\n", challenge_3_text)

					m, err := cryptor.Int(cryptor.Reader, new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil))

					if err != nil {
						fmt.Fprintln(os.Stdout, "Unable to generate random message!")
						os.Exit(1);
					}

					fmt.Fprint(os.Stdout, "\nWhat is n for your public key? ")
					ok := input.Scan()
					if !ok {
						fmt.Fprintln(os.Stdout, "Error reading input!")
						break
					}

					n, ok := new(big.Int).SetString(input.Text(), 10)

					if !ok || n == nil {
						fmt.Fprintln(os.Stdout, "Error parsing n!")
						break
					}

					if n.Cmp(new(big.Int).Exp(big.NewInt(2), big.NewInt(512), nil)) < 0 {
						fmt.Fprintln(os.Stdout, "n is too small, must be at least 2^512!")
						break
					}

					if n.ProbablyPrime(32) == true {
						fmt.Fprintln(os.Stdout, "n must be composite!")
						break
					}

					fmt.Fprint(os.Stdout, "\nWhat is e for your public key? ")
					ok = input.Scan()
					if !ok {
						fmt.Fprintln(os.Stdout, "Error reading input!")
						break
					}

					e, ok := new(big.Int).SetString(input.Text(), 10)

					if !ok || e == nil {
						fmt.Fprintln(os.Stdout, "Error parsing e!")
						break
					}

					if n.Cmp(big.NewInt(100)) < 0 {
						fmt.Fprintln(os.Stdout, "e is too small, must be at least 100!")
						break
					}

					c := new(big.Int).Exp(m, e, n)

					fmt.Fprintf(os.Stdout, "The ciphertext you must decrypt is %s\n", c.Text(10))
					fmt.Fprint(os.Stdout, "\nWhat was m (decrypt with the d you've made for your key!)? ")
					ok = input.Scan()
					if !ok {
						fmt.Fprintln(os.Stdout, "Error reading input!")
						break
					}

					userm, ok := new(big.Int).SetString(input.Text(), 10)

					if !ok || userm == nil {
						fmt.Fprintln(os.Stdout, "Unable to parse provided m!")
						break
					}

					if m.Cmp(userm) != 0 {
						fmt.Fprintf(os.Stdout, "Unfortunately that is not correct.\n")
						fmt.Fprintf(os.Stdout, "The correct answer was %s\n", m.Text(10))
						fmt.Fprintf(os.Stdout, "c was generated with you public key by c = m ^ e mod n:\n")
						fmt.Fprintf(os.Stdout, "%s = %s ^ %s mod %s\n\n", c.Text(10), m.Text(10), e.Text(10), n.Text(10))
						fmt.Fprintf(os.Stdout, "You should double-check that your key was generated correctly\n")
						fmt.Fprintf(os.Stdout, "Especially check that e * n mod phi(n) = 1\n")
						fmt.Fprintf(os.Stdout, "And check that 12345 ^ (e * d) mod n = 12345\n")

						break
					} else {
						fmt.Fprintf(os.Stdout, "Correct!")
						challenge_status[2] = true
					}


				case "4": // =========================== 4 ============================

					fmt.Fprintf(os.Stdout, "Challenge 4:\n%s\n", challenge_4_text)

					fails := 0
				ch_4_retry_key:
					p, err := cryptor.Prime(cryptor.Reader, 64)

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate prime p\n")
						os.Exit(1)
					}

					q, err := cryptor.Prime(cryptor.Reader, 64)

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate prime q\n")
						os.Exit(1)
					}

					n := new(big.Int).Mul(p, q)

					pm1 := new(big.Int).Add(p, big.NewInt(-1))
					qm1 := new(big.Int).Add(q, big.NewInt(-1))

					etot := new(big.Int).Mul(pm1, qm1)

					e := big.NewInt(65537)
					d := new(big.Int).ModInverse(e, etot)

					if d == nil {
						if (fails > 5) {
							fmt.Fprintf(os.Stderr, "Error: unable to generate d! Probably (p - 1) or (q - 1) was a multiple of e\n")
							os.Exit(1)
						} else {
							fails++
							goto ch_4_retry_key
						}
					}

					m, err := cryptor.Int(cryptor.Reader, new(big.Int).Exp(big.NewInt(2), big.NewInt(96), nil))

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate m\n")
						os.Exit(1)
					}

					c := new(big.Int).Exp(m, e, n)

					fmt.Fprintf(os.Stdout, "== Generated Public Key ==\n")
					fmt.Fprintf(os.Stdout, "Public Modulus (n): %s\n", n.Text(10))
					fmt.Fprintf(os.Stdout, "Encrypting Exponent (e): %s\n", e.Text(10))
					fmt.Fprintf(os.Stdout, "\n")
					fmt.Fprintf(os.Stdout, "Message to decrypt (c): %s\n", c.Text(10))

					fmt.Fprint(os.Stdout, "\nWhat is the plaintext (m) for the encrypted message (c) for the public key? ")
					ok := input.Scan()
					if !ok {
						fmt.Fprintln(os.Stdout, "Error reading input!")
						break
					}

					mtext, ok := new(big.Int).SetString(input.Text(), 10)

					if !ok || mtext == nil {
						fmt.Fprintln(os.Stdout, "Error parsing answer!")
						break
					}

					if mtext.Cmp(m) == 0 {
						fmt.Fprintf(os.Stdout, "Correct!")
						challenge_status[3] = true
					} else {
						fmt.Fprintf(os.Stdout, "Unfortunately that is not correct.\n")
						fmt.Fprintf(os.Stdout, "The correct answer was %s\n", m.Text(10))
						fmt.Fprintf(os.Stdout, "Which could be found by first factoring n into p and q:\n")
						fmt.Fprintf(os.Stdout, "%s * %s = %s\n", p.Text(10), q.Text(10), n.Text(10))
						fmt.Fprintf(os.Stdout, "And then finding the inverse of e mod phi(n):\n")
						fmt.Fprintf(os.Stdout, "%s * %s mod %s = 1\n", e.Text(10), d.Text(10), etot.Text(10))
						fmt.Fprintf(os.Stdout, "Finally, m could be found with c^d mod n:\n")
						fmt.Fprintf(os.Stdout, "%s ^ %s mod %s\n = %s\n\n", c.Text(10), d.Text(10), n.Text(10), m.Text(10))
						fmt.Fprintf(os.Stdout, "Did you successfully factor n into p and q?\n")
					}

				case "5": // =========================== 5 ============================

					fmt.Fprintf(os.Stdout, "Challenge 5:\n%s\n", challenge_5_text)

					fails := 0
				ch_5_retry_key:
					p, err := cryptor.Prime(cryptor.Reader, 1024)

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate prime p\n")
						os.Exit(1)
					}


					delta, err := cryptor.Int(cryptor.Reader, new(big.Int).Exp(big.NewInt(2), big.NewInt(524), nil))

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to delta for q\n")
						os.Exit(1)
					}

					q := new(big.Int).Add(p, delta)

					for {
						if q.ProbablyPrime(32) == true {
							break
						} else {
							q.Add(q, big.NewInt(1))
						}
					}

					n := new(big.Int).Mul(p, q)

					pm1 := new(big.Int).Add(p, big.NewInt(-1))
					qm1 := new(big.Int).Add(q, big.NewInt(-1))

					etot := new(big.Int).Mul(pm1, qm1)

					e := big.NewInt(65537)
					d := new(big.Int).ModInverse(e, etot)

					if d == nil {
						if (fails > 5) {
							fmt.Fprintf(os.Stderr, "Error: unable to generate d! Probably (p - 1) or (q - 1) was a multiple of e\n")
							os.Exit(1)
						} else {
							fails++
							goto ch_5_retry_key
						}
					}

					m, err := cryptor.Int(cryptor.Reader, new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil))

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate m\n")
						os.Exit(1)
					}

					c := new(big.Int).Exp(m, e, n)

					fmt.Fprintf(os.Stdout, "== Generated Public Key ==\n")
					fmt.Fprintf(os.Stdout, "Public Modulus (n): %s\n", n.Text(10))
					fmt.Fprintf(os.Stdout, "Encrypting Exponent (e): %s\n", e.Text(10))
					fmt.Fprintf(os.Stdout, "\n")
					fmt.Fprintf(os.Stdout, "Message to decrypt (c): %s\n", c.Text(10))

					fmt.Fprint(os.Stdout, "\nWhat is the plaintext (m) for the encrypted message (c) for the public key? ")
					ok := input.Scan()
					if !ok {
						fmt.Fprintln(os.Stdout, "Error reading input!")
						break
					}

					mtext, ok := new(big.Int).SetString(input.Text(), 10)

					if !ok || mtext == nil {
						fmt.Fprintln(os.Stdout, "Error parsing answer!")
						break
					}

					// s = sqrtint(n); for(X = 1, 2^26, if(issquare((s + X)^2 - n) == 1, printf("X delta: %d\n", X); break()))
					// # l is the limit to search
					// fermat_factor(n, l) = {my(s); s = sqrtint(n) + 1; for(X = 0, l, if(issquare((s + X)^2 - n) == 1, return((s + X) - sqrtint((s + X)^2 - n))))}

					if mtext.Cmp(m) == 0 {
						fmt.Fprintf(os.Stdout, "Correct!")
						challenge_status[4] = true
					} else {
						fmt.Fprintf(os.Stdout, "Unfortunately that is not correct.\n")
						fmt.Fprintf(os.Stdout, "The correct answer was %s\n", m.Text(10))
						fmt.Fprintf(os.Stdout, "Which could be found by first factoring n into p and q:\n")
						fmt.Fprintf(os.Stdout, "%s * %s = %s\n", p.Text(10), q.Text(10), n.Text(10))
						fmt.Fprintf(os.Stdout, "And then finding the inverse of e mod phi(n):\n")
						fmt.Fprintf(os.Stdout, "%s * %s mod %s = 1\n", e.Text(10), d.Text(10), etot.Text(10))
						fmt.Fprintf(os.Stdout, "Finally, m could be found with c^d mod n:\n")
						fmt.Fprintf(os.Stdout, "%s ^ %s mod %s\n = %s\n\n", c.Text(10), d.Text(10), n.Text(10), m.Text(10))
						fmt.Fprintf(os.Stdout, "Did you successfully factor n into p and q?\n")
					}

				case "6": // =========================== 6 ============================

					fmt.Fprintf(os.Stdout, "Challenge 6:\n%s\n", challenge_6_text)

					fails := 0
				ch_6_retry_key:
					p, err := cryptor.Prime(cryptor.Reader, 1024)

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate prime p\n")
						os.Exit(1)
					}

					q1, err := cryptor.Prime(cryptor.Reader, 1024)

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate prime q1\n")
						os.Exit(1)
					}

					q2, err := cryptor.Prime(cryptor.Reader, 1024)

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate prime q2\n")
						os.Exit(1)
					}

					n1 := new(big.Int).Mul(p, q1)
					n2 := new(big.Int).Mul(p, q2)

					pm1 := new(big.Int).Add(p, big.NewInt(-1))
					q1m1 := new(big.Int).Add(q1, big.NewInt(-1))
					q2m1 := new(big.Int).Add(q2, big.NewInt(-1))

					etot1 := new(big.Int).Mul(pm1, q1m1)
					etot2 := new(big.Int).Mul(pm1, q2m1)

					e := big.NewInt(65537)

					d1 := new(big.Int).ModInverse(e, etot1)

					d2 := new(big.Int).ModInverse(e, etot2)

					if d1 == nil || d2 == nil {
						if (fails > 5) {
							fmt.Fprintf(os.Stderr, "Error: unable to generate d! Probably (p - 1) or (q - 1) was a multiple of e\n")
							os.Exit(1)
						} else {
							fails++
							goto ch_6_retry_key
						}
					}

					m, err := cryptor.Int(cryptor.Reader, new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil))

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate m\n")
						os.Exit(1)
					}

					c1 := new(big.Int).Exp(m, e, n1)
					c2 := new(big.Int).Exp(m, e, n2)

					fmt.Fprintf(os.Stdout, "== Generated Public Key 1 ==\n")
					fmt.Fprintf(os.Stdout, "Public Modulus (n): %s\n", n1.Text(10))
					fmt.Fprintf(os.Stdout, "Encrypting Exponent (e): %s\n", e.Text(10))
					fmt.Fprintf(os.Stdout, "Encrypted message (c): %s\n", c1.Text(10))

					fmt.Fprintf(os.Stdout, "\n== Generated Public Key 2 ==\n")
					fmt.Fprintf(os.Stdout, "Public Modulus (n): %s\n", n2.Text(10))
					fmt.Fprintf(os.Stdout, "Encrypting Exponent (e): %s\n", e.Text(10))
					fmt.Fprintf(os.Stdout, "Encrypted message (c): %s\n", c2.Text(10))

					fmt.Fprint(os.Stdout, "\nThe same message (m) was encrypted with both keys, what was the message? ")
					ok := input.Scan()
					if !ok {
						fmt.Fprintln(os.Stdout, "Error reading input!")
						break
					}

					mtext, ok := new(big.Int).SetString(input.Text(), 10)

					if !ok || mtext == nil {
						fmt.Fprintln(os.Stdout, "Error parsing answer!")
						break
					}

					if mtext.Cmp(m) == 0 {
						fmt.Fprintf(os.Stdout, "Correct!")
						challenge_status[5] = true
					} else {
						fmt.Fprintf(os.Stdout, "Unfortunately that is not correct.\n")
						fmt.Fprintf(os.Stdout, "The correct answer was %s\n", m.Text(10))
						fmt.Fprintf(os.Stdout, "Which could be found by first finding the common prime:\n")
						fmt.Fprintf(os.Stdout, "GCD(%d, %d) = %s\n", n1.Text(10), n2.Text(10), p.Text(10))
						fmt.Fprintf(os.Stdout, "Did you successfully find the common factor?\n")
					}

				case "7": // =========================== 7 ============================

					fmt.Fprintf(os.Stdout, "Challenge 7:\n%s\n", challenge_7_text)

					fails := 0
				ch_7_retry_key:

					var p, q, n, d, pm1, qm1, etot, c [4]*big.Int
					var err error

					e := big.NewInt(101)

					m, err := cryptor.Int(cryptor.Reader, new(big.Int).Exp(big.NewInt(2), big.NewInt(64), nil))

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate m\n")
						os.Exit(1)
					}

					for i := 0; i < 4; i++ {

						p[i], err = cryptor.Prime(cryptor.Reader, 1024)

						if err != nil {
							fmt.Fprintf(os.Stderr, "Error: unable to generate prime p\n")
							os.Exit(1)
						}

						q[i], err = cryptor.Prime(cryptor.Reader, 1024)

						if err != nil {
							fmt.Fprintf(os.Stderr, "Error: unable to generate prime q\n")
							os.Exit(1)
						}

						n[i] = new(big.Int).Mul(p[i], q[i])

						pm1[i] = new(big.Int).Add(p[i], big.NewInt(-1))
						qm1[i] = new(big.Int).Add(q[i], big.NewInt(-1))

						etot[i] = new(big.Int).Mul(pm1[i], qm1[i])

						d[i] = new(big.Int).ModInverse(e, etot[i])

						if d[i] == nil {
							if (fails > 5) {
								fmt.Fprintf(os.Stderr, "Error: unable to generate d! Probably (p - 1) or (q - 1) was a multiple of e\n")
								os.Exit(1)
							} else {
								fails++
								goto ch_7_retry_key
							}
						}

						c[i] = new(big.Int).Exp(m, e, n[i])
					}

					for i := 0; i < 4; i++ {
						fmt.Fprintf(os.Stdout, "\n== Generated Public Key %d ==\n", i + 1)
						fmt.Fprintf(os.Stdout, "Public Modulus (n): %s\n", n[i].Text(10))
						fmt.Fprintf(os.Stdout, "Encrypting Exponent (e): %s\n", e.Text(10))
						fmt.Fprintf(os.Stdout, "Encrypted message (c): %s\n", c[i].Text(10))
					}


					fmt.Fprint(os.Stdout, "\nThe same message (m) was encrypted with all keys, what was the message? ")
					ok := input.Scan()
					if !ok {
						fmt.Fprintln(os.Stdout, "Error reading input!")
						break
					}

					mtext, ok := new(big.Int).SetString(input.Text(), 10)

					if !ok || mtext == nil {
						fmt.Fprintln(os.Stdout, "Error parsing answer!")
						break
					}

					if mtext.Cmp(m) == 0 {
						fmt.Fprintf(os.Stdout, "Correct!")
						challenge_status[6] = true
					} else {
						fmt.Fprintf(os.Stdout, "Unfortunately that is not correct.\n")
						fmt.Fprintf(os.Stdout, "Did you find the unique value for m^e mod (n1 * n2 * n3 * n4)?\n")
						fmt.Fprintf(os.Stdout, "Did you take the 101st integer root?\n")
					}

				case "8": // =========================== 8 ============================

					fmt.Fprintf(os.Stdout, "Challenge 8:\n%s\n", challenge_8_text)

					// x, m := CRT(big.NewInt(7), big.NewInt(51), big.NewInt(2), big.NewInt(101))
					// fmt.Fprintf(os.Stdout, "Mod(%s, %s)\n", x.Text(10), m.Text(10))

					fails := 0
				ch_8_retry_key:
					p, err := cryptor.Prime(cryptor.Reader, 1024)

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate prime p\n")
						os.Exit(1)
					}

					q, err := cryptor.Prime(cryptor.Reader, 1024)

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate prime q\n")
						os.Exit(1)
					}

					n := new(big.Int).Mul(p, q)

					pm1 := new(big.Int).Add(p, big.NewInt(-1))
					qm1 := new(big.Int).Add(q, big.NewInt(-1))

					etot := new(big.Int).Mul(pm1, qm1)

					e := big.NewInt(65537)
					d := new(big.Int).ModInverse(e, etot)

					if d == nil {
						if (fails > 5) {
							fmt.Fprintf(os.Stderr, "Error: unable to generate d! Probably (p - 1) or (q - 1) was a multiple of e\n")
							os.Exit(1)
						} else {
							fails++
							goto ch_8_retry_key
						}
					}

					d_pm1 := new(big.Int).Mod(d, pm1)
					d_qm1 := new(big.Int).Mod(d, qm1)

					m, err := cryptor.Int(cryptor.Reader, new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil))

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate m\n")
						os.Exit(1)
					}

					mtest, err := cryptor.Int(cryptor.Reader, new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil))

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate mtest\n")
						os.Exit(1)
					}

					ctest := new(big.Int).Exp(mtest, e, n) // Used to test answer

					s := new(big.Int).Exp(m, d, n)
					//s_crt, n_crt := CRT(new(big.Int).Exp(m, d_pm1, p), p, new(big.Int).Exp(m, d_qm1, q), q)

					// Adding 1 here is the error
					s_crt_error, _ := CRT(new(big.Int).Add(new(big.Int).Exp(m, d_pm1, p), big.NewInt(1)), p, new(big.Int).Exp(m, d_qm1, q), q)

					//fmt.Fprintf(os.Stdout, "CRT-found Modulus (n): %s\n", n_crt.Text(10))

					fmt.Fprintf(os.Stdout, "== Generated Public Key ==\n")
					fmt.Fprintf(os.Stdout, "Public Modulus (n): %s\n", n.Text(10))
					fmt.Fprintf(os.Stdout, "Encrypting Exponent (e): %s\n", e.Text(10))
					fmt.Fprintf(os.Stdout, "\nSigned Message (m): %s\n", m.Text(10))
					fmt.Fprintf(os.Stdout, "\nGood Message Signature (s): %s\n", s.Text(10))
					fmt.Fprintf(os.Stdout, "\nBad Message Signature (s): %s\n", s_crt_error.Text(10))
					fmt.Fprintf(os.Stdout, "\n")


					fmt.Fprint(os.Stdout, "\nWhat is the private decrypting exponent (d) for the public key? ")
					ok := input.Scan()
					if !ok {
						fmt.Fprintln(os.Stdout, "Error reading input!")
						break
					}

					dtext, ok := new(big.Int).SetString(input.Text(), 10)

					if !ok || dtext == nil {
						fmt.Fprintln(os.Stdout, "Error parsing answer!")
						break
					}

					mdec := new(big.Int).Exp(ctest, dtext, n) // test answer

					if mdec.Cmp(mtest) == 0 {
						fmt.Fprintf(os.Stdout, "Correct!")
						challenge_status[7] = true
					} else {
						fmt.Fprintf(os.Stdout, "Unfortunately that is not correct.\n")
						fmt.Fprintf(os.Stdout, "The correct answer was %s\n", d.Text(10))
					}

				case "9": // =========================== 9 ============================

					fmt.Fprintf(os.Stdout, "Challenge 9:\n%s\n", challenge_9_text)

					fails := 0
				ch_9_retry_key:
					p, err := cryptor.Prime(cryptor.Reader, 1024)

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate prime p\n")
						os.Exit(1)
					}

					q, err := cryptor.Prime(cryptor.Reader, 1024)

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate prime q\n")
						os.Exit(1)
					}

					n := new(big.Int).Mul(p, q)

					pm1 := new(big.Int).Add(p, big.NewInt(-1))
					qm1 := new(big.Int).Add(q, big.NewInt(-1))

					etot := new(big.Int).Mul(pm1, qm1)

					e := big.NewInt(65537)
					d := new(big.Int).ModInverse(e, etot)

					if d == nil {
						if (fails > 5) {
							fmt.Fprintf(os.Stderr, "Error: unable to generate d! Probably (p - 1) or (q - 1) was a multiple of e\n")
							os.Exit(1)
						} else {
							fails++
							goto ch_9_retry_key
						}
					}

					fmt.Fprintf(os.Stdout, "== Generated Public Key ==\n")
					fmt.Fprintf(os.Stdout, "Public Modulus (n): %s\n", n.Text(10))
					fmt.Fprintf(os.Stdout, "\nPublic Exponent (e): %s\n", e.Text(10))

					fmt.Fprintf(os.Stdout, "\n")


					fmt.Fprint(os.Stdout, "\n\"Prove\" that you know the private key by providing a message and its signature\n")
					fmt.Fprint(os.Stdout, "\nProvide a message: ")
					ok := input.Scan()
					if !ok {
						fmt.Fprintln(os.Stdout, "Error reading input!")
						break
					}

					m, ok := new(big.Int).SetString(input.Text(), 10)

					if !ok {
						fmt.Fprintln(os.Stdout, "Error reading input!")
						break
					}

					if m.Cmp(big.NewInt(2)) < 0 {
						fmt.Fprint(os.Stdout, "\nNice try but m must be at least 2")
						break
					}

					fmt.Fprint(os.Stdout, "\nNow provide the message signature: ")
					ok = input.Scan()
					if !ok {
						fmt.Fprintln(os.Stdout, "Error reading input!")
						break
					}

					s, ok := new(big.Int).SetString(input.Text(), 10)

					if !ok {
						fmt.Fprintln(os.Stdout, "Error reading input!")
						break
					}

					// Compute the real sig
					rs := new(big.Int).Exp(m, d, n)

					if rs.Cmp(s) == 0 {
						fmt.Fprintf(os.Stdout, "Correct!")
						challenge_status[8] = true
					} else {
						fmt.Fprintf(os.Stdout, "Unfortunately that is not correct.\n")
						fmt.Fprintf(os.Stdout, "Remember, to generate a backwards signature, choose the signature and encrypt it to find the message.\n");
					}

				case "10": // =========================== 10 ============================

					fmt.Fprintf(os.Stdout, "Challenge 10:\n%s\n", challenge_10_text)

					// x, m := CRT(big.NewInt(7), big.NewInt(51), big.NewInt(2), big.NewInt(101))
					// fmt.Fprintf(os.Stdout, "Mod(%s, %s)\n", x.Text(10), m.Text(10))

					fails := 0
				ch_10_retry_key:
					p, err := cryptor.Prime(cryptor.Reader, 1024)

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate prime p\n")
						os.Exit(1)
					}

					q, err := cryptor.Prime(cryptor.Reader, 1024)

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate prime q\n")
						os.Exit(1)
					}

					n := new(big.Int).Mul(p, q)

					pm1 := new(big.Int).Add(p, big.NewInt(-1))
					qm1 := new(big.Int).Add(q, big.NewInt(-1))

					etot := new(big.Int).Mul(pm1, qm1)

					e := big.NewInt(65537)
					d := new(big.Int).ModInverse(e, etot)

					if d == nil {
						if (fails > 5) {
							fmt.Fprintf(os.Stderr, "Error: unable to generate d! Probably (p - 1) or (q - 1) was a multiple of e\n")
							os.Exit(1)
						} else {
							fails++
							goto ch_10_retry_key
						}
					}

					// Generate the 256 bit "hash" message in three parts of size 120, 80, and 56 bits to make it easily factorable
					h_120, err := cryptor.Int(cryptor.Reader, new(big.Int).Exp(big.NewInt(2), big.NewInt(120), nil))

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate h_120\n")
						os.Exit(1)
					}

					h_80, err := cryptor.Int(cryptor.Reader, new(big.Int).Exp(big.NewInt(2), big.NewInt(80), nil))

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate h_80\n")
						os.Exit(1)
					}

					h_56, err := cryptor.Int(cryptor.Reader, new(big.Int).Exp(big.NewInt(2), big.NewInt(56), nil))

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate h_56\n")
						os.Exit(1)
					}

					h := new(big.Int).Mul(new(big.Int).Mul(h_120, h_80), h_56)

					s := new(big.Int).Exp(h, d, n)


					fmt.Fprintf(os.Stdout, "== Generated Public Key ==\n")
					fmt.Fprintf(os.Stdout, "Public Modulus (n): %s\n", n.Text(10))
					fmt.Fprintf(os.Stdout, "\nHash to be signed (h): %s\n", h.Text(10))

					fmt.Fprintf(os.Stdout, "\n")

					for {
						fmt.Fprint(os.Stdout, "\nEnter a small 'hash' to be signed (limit 128 bits) (enter 0 to end signing loop): ")
						ok := input.Scan()
						if !ok {
							fmt.Fprintln(os.Stdout, "Error reading input!")
							continue
						}

						sm, ok := new(big.Int).SetString(input.Text(), 10)

						if !ok {
							fmt.Fprintln(os.Stdout, "Error reading input!")
							continue
						}

						if sm.Cmp(big.NewInt(0)) == 0 {
							break
						}

						if sm.Cmp(big.NewInt(0)) < 0 {
							fmt.Fprintln(os.Stdout, "Message to sign must be positive!")
							continue
						}

						if sm.Cmp(new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil)) > 0 {
							fmt.Fprintln(os.Stdout, "Message to sign must be less than 128 bits!")
							continue
						}

						sms := new(big.Int).Exp(sm, d, n)

						fmt.Fprintf(os.Stdout, "\nSignature: %s\n", sms.Text(10))
					}

					fmt.Fprint(os.Stdout, "\nWhat is the signature for the hash to be signed? ")
					ok := input.Scan()
					if !ok {
						fmt.Fprintln(os.Stdout, "Error reading input!")
						break
					}

					hs, ok := new(big.Int).SetString(input.Text(), 10)

					if !ok {
						fmt.Fprintln(os.Stdout, "Error reading input!")
						break
					}

					if hs.Cmp(s) == 0 {
						fmt.Fprintf(os.Stdout, "Correct!")
						challenge_status[9] = true
					} else {
						fmt.Fprintf(os.Stdout, "Unfortunately that is not correct.\n")
						fmt.Fprintf(os.Stdout, "The correct answer was %s\n", s.Text(10))
						fmt.Fprintf(os.Stdout, "Did you factor the hash to be signed into small chunks?\n");
						fmt.Fprintf(os.Stdout, "Factor the hash, sign smaller chunks, and then construct a full signature from them.\n");
					}


				case "11": // =========================== 11 ============================

					fmt.Fprintf(os.Stdout, "Challenge 11:\n%s\n", challenge_11_text)

					// x, m := CRT(big.NewInt(7), big.NewInt(51), big.NewInt(2), big.NewInt(101))
					// fmt.Fprintf(os.Stdout, "Mod(%s, %s)\n", x.Text(10), m.Text(10))

					fails := 0
				ch_11_retry_key:

					// First we need to generate a weak b-smooth p - 1
					// Start with an initial p - 1 to save a lot of work when we retry a bunch
					ipm1 := big.NewInt(2)
					for i := 0; i < (1024 / 16) - 3; i++ { // Stop about 3 short
						sprime, err := cryptor.Prime(cryptor.Reader, 16)

						if err != nil {
							fmt.Fprintf(os.Stderr, "Error: unable to generate small prime\n")
							os.Exit(1)
						}

						ipm1 = new(big.Int).Mul(ipm1, sprime)
					}

					// Now ipm1 is almost big enough, just mulitply ~3 more small primes with it
					// until ip1 + 1 happens to be prime

					var p *big.Int
					pfound := false
					for !pfound {
						pm1 := ipm1

						for pm1.Cmp(new(big.Int).Exp(big.NewInt(2), big.NewInt(1024), nil)) < 0 {
							sprime, err := cryptor.Prime(cryptor.Reader, 16)

							if err != nil {
								fmt.Fprintf(os.Stderr, "Error: unable to generate small prime\n")
								os.Exit(1)
							}

							pm1 = new(big.Int).Mul(pm1, sprime)
						}

						p = new(big.Int).Add(pm1, big.NewInt(1))
						if p.ProbablyPrime(32) == true {
							pfound = true
							break;
						}
					}

					q, err := cryptor.Prime(cryptor.Reader, 1024)

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate prime q\n")
						os.Exit(1)
					}

					n := new(big.Int).Mul(p, q)

					pm1 := new(big.Int).Add(p, big.NewInt(-1))
					qm1 := new(big.Int).Add(q, big.NewInt(-1))

					etot := new(big.Int).Mul(pm1, qm1)

					e := big.NewInt(65537)
					d := new(big.Int).ModInverse(e, etot)

					if d == nil {
						if (fails > 5) {
							fmt.Fprintf(os.Stderr, "Error: unable to generate d! Probably (p - 1) or (q - 1) was a multiple of e\n")
							os.Exit(1)
						} else {
							fails++
							goto ch_11_retry_key
						}
					}

					m, err := cryptor.Int(cryptor.Reader, new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil))

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: unable to generate m\n")
						os.Exit(1)
					}

					c := new(big.Int).Exp(m, e, n)

					fmt.Fprintf(os.Stdout, "== Generated Public Key ==\n")
					fmt.Fprintf(os.Stdout, "Public Modulus (n): %s\n", n.Text(10))
					fmt.Fprintf(os.Stdout, "Public Exponent (e): %s\n", e.Text(10))
					fmt.Fprintf(os.Stdout, "Encrypted Message (c): %s\n", c.Text(10))

					fmt.Fprintf(os.Stdout, "\n")

					fmt.Fprint(os.Stdout, "\nWhat was the message that was encrypted? ")
					ok := input.Scan()
					if !ok {
						fmt.Fprintln(os.Stdout, "Error reading input!")
						break
					}

					um, ok := new(big.Int).SetString(input.Text(), 10)

					if !ok {
						fmt.Fprintln(os.Stdout, "Error reading input!")
						break
					}

					if um.Cmp(m) == 0 {
						fmt.Fprintf(os.Stdout, "Correct!")
						challenge_status[10] = true
					} else {
						fmt.Fprintf(os.Stdout, "Unfortunately that is not correct.\n")
						fmt.Fprintf(os.Stdout, "The correct answer was %s\n", m.Text(10))
						fmt.Fprintf(os.Stdout, "Did you factor n by exploiting the fact that p - 1 has many small factors?\n");
						fmt.Fprintf(os.Stdout, "p - 1 is 65535-smooth so you only have to try primes up to 65521\n");
					}

					// =========================== END ============================
				default:
					fmt.Fprintf(os.Stdout, "\"%s\" argument not understood. challenge command requires one numeric argument. Try \"help\" for a list of commands.", tokens[1])
				}

			} else {
				fmt.Fprintf(os.Stdout, "challenge command requires one numeric argument. Try \"help\" for a list of commands.")
			}

		case "getflag":
			if len(tokens) == 2 {
				switch tokens[1] {

				case "2022-first":
					if challenge_status[0] {
						fmt.Fprintf(os.Stdout, "\nChallenge 1, %s, complete!\n", challenge_1_name)
					} else {
						fmt.Fprintf(os.Stdout, "\nChallenge 1 incomplete!\n")
						break
					}

					if challenge_status[1] {
						fmt.Fprintf(os.Stdout, "\nChallenge 2, %s, complete!\n", challenge_2_name)
					} else {
						fmt.Fprintf(os.Stdout, "\nChallenge 2 incomplete!\n")
						break
					}

					if challenge_status[2] {
						fmt.Fprintf(os.Stdout, "\nChallenge 3, %s, complete!\n", challenge_3_name)
					} else {
						fmt.Fprintf(os.Stdout, "\nChallenge 3 incomplete!\n")
						break
					}

					if challenge_status[3] {
						fmt.Fprintf(os.Stdout, "\nChallenge 4 complete!\n")
					} else {
						fmt.Fprintf(os.Stdout, "\nChallenge 4 incomplete!\n")
						break
					}

					fmt.Fprintf(os.Stdout, "\n2022-first set complete!\n%s\n", flag_2022_first)

				case "2022-second":
					if challenge_status[4] {
						fmt.Fprintf(os.Stdout, "\nChallenge 5 complete!\n")
					} else {
						fmt.Fprintf(os.Stdout, "\nChallenge 5 incomplete!\n")
						break
					}

					if challenge_status[5] {
						fmt.Fprintf(os.Stdout, "\nChallenge 6 complete!\n")
					} else {
						fmt.Fprintf(os.Stdout, "\nChallenge 6 incomplete!\n")
						break
					}

					if challenge_status[6] {
						fmt.Fprintf(os.Stdout, "\nChallenge 7 complete!\n")
					} else {
						fmt.Fprintf(os.Stdout, "\nChallenge 7 incomplete!\n")
						break
					}

					if challenge_status[7] {
						fmt.Fprintf(os.Stdout, "\nChallenge 8 complete!\n")
					} else {
						fmt.Fprintf(os.Stdout, "\nChallenge 8 incomplete!\n")
						break
					}

					fmt.Fprintf(os.Stdout, "\n2022-second set complete!\n%s\n", flag_2022_second)

				case "2023":
					if challenge_status[8] {
						fmt.Fprintf(os.Stdout, "\nChallenge 9 complete!\n")
					} else {
						fmt.Fprintf(os.Stdout, "\nChallenge 9 incomplete!\n")
						break
					}

					if challenge_status[9] {
						fmt.Fprintf(os.Stdout, "\nChallenge 10 complete!\n")
					} else {
						fmt.Fprintf(os.Stdout, "\nChallenge 10 incomplete!\n")
						break
					}

					if challenge_status[10] {
						fmt.Fprintf(os.Stdout, "\nChallenge 11 complete!\n")
					} else {
						fmt.Fprintf(os.Stdout, "\nChallenge 11 incomplete!\n")
						break
					}

					fmt.Fprintf(os.Stdout, "\n2023 set complete!\n%s\n", flag_2023)

				default:
					fmt.Fprintf(os.Stdout, "\"%s\" argument not understood. attack command requires one of {2022-first, 2022-second, 2023}. Try \"help\" for a list of commands.", tokens[1])

				}

			} else {
				fmt.Fprintf(os.Stdout, "getflag command requires one argument. Try \"help\" for a list of commands.")
			}

		case "status":
			print_status()

		case "h":
			print_help()

		case "?":
			print_help()

		case "exit":
			exit = true

		case "quit":
			exit = true

		case "flag":
			fmt.Fprintf(os.Stdout, "lolz you typed 'flag' but that isn't a command. You didn't really think that was going to work, did you?\n")
			exit = true

		case "^d":
			fmt.Fprintf(os.Stdout, "Uhmmm... You do realize that the '^' in '^d' isn't a literal '^' right??")

		default:
			fmt.Fprintf(os.Stdout, "%s: `%s` command not found. Try \"help\" for a list of commands.", prog_name, tokens[0])

		}
	}

}


func print_status() {
	fmt.Fprintf(os.Stdout, "\nChallenge Statuses\n")
	fmt.Fprintf(os.Stdout, "==================\n")

	fmt.Fprintf(os.Stdout, "\n2022-first:\n")

	fmt.Fprintf(os.Stdout, " 1 - ")
	if challenge_status[0] {
		fmt.Fprintf(os.Stdout, "<complete> ")
	}
	fmt.Fprintf(os.Stdout, "%s\n", challenge_1_name)

	fmt.Fprintf(os.Stdout, " 2 - ")
	if challenge_status[1] {
		fmt.Fprintf(os.Stdout, "<complete> ")
	}
	fmt.Fprintf(os.Stdout, "%s\n", challenge_2_name)

	fmt.Fprintf(os.Stdout, " 3 - ")
	if challenge_status[2] {
		fmt.Fprintf(os.Stdout, "<complete> ")
	}
	fmt.Fprintf(os.Stdout, "%s\n", challenge_3_name)

	fmt.Fprintf(os.Stdout, " 4 - ")
	if challenge_status[3] {
		fmt.Fprintf(os.Stdout, "<complete> ")
	}
	fmt.Fprintf(os.Stdout, "%s\n", challenge_4_name)


	fmt.Fprintf(os.Stdout, "\n2022-second:\n")

	fmt.Fprintf(os.Stdout, " 5 - ")
	if challenge_status[4] {
		fmt.Fprintf(os.Stdout, "<complete> ")
	}
	fmt.Fprintf(os.Stdout, "%s\n", challenge_5_name)

	fmt.Fprintf(os.Stdout, " 6 - ")
	if challenge_status[5] {
		fmt.Fprintf(os.Stdout, "<complete> ")
	}
	fmt.Fprintf(os.Stdout, "%s\n", challenge_6_name)

	fmt.Fprintf(os.Stdout, " 7 - ")
	if challenge_status[6] {
		fmt.Fprintf(os.Stdout, "<complete> ")
	}
	fmt.Fprintf(os.Stdout, "%s\n", challenge_7_name)

	fmt.Fprintf(os.Stdout, " 8 - ")
	if challenge_status[7] {
		fmt.Fprintf(os.Stdout, "<complete> ")
	}
	fmt.Fprintf(os.Stdout, "%s\n", challenge_8_name)


	fmt.Fprintf(os.Stdout, "\n2023:\n")

	fmt.Fprintf(os.Stdout, " 9 - ")
	if challenge_status[8] {
		fmt.Fprintf(os.Stdout, "<complete> ")
	}
	fmt.Fprintf(os.Stdout, "%s\n", challenge_9_name)

	fmt.Fprintf(os.Stdout, " 10 - ")
	if challenge_status[9] {
		fmt.Fprintf(os.Stdout, "<complete> ")
	}
	fmt.Fprintf(os.Stdout, "%s\n", challenge_10_name)

	fmt.Fprintf(os.Stdout, " 11 - ")
	if challenge_status[10] {
		fmt.Fprintf(os.Stdout, "<complete> ")
	}
	fmt.Fprintf(os.Stdout, "%s\n", challenge_11_name)

}


func print_help() {
	fmt.Fprintf(os.Stdout, "\n%s help:\n%s", prog_name, help_text)
}


func print_sample_code() {
	fmt.Fprint(os.Stdout, "\n// === START SAMPLE CODE ===\n")
	fmt.Fprintf(os.Stdout, "%s", sample_code)
	fmt.Fprint(os.Stdout, "\n// === END SAMPLE CODE ===\n")
}


func rand_bytes(n int) []byte {

	b := make([]byte, n)

	_, err := cryptor.Read(b)

	if err != nil {
		os.Exit(-1)
	}

	return b
}

func CRT(r1, m1, r2, m2 *big.Int) (*big.Int, *big.Int) {

	x := new(big.Int)
	y := new(big.Int)

	// Extended euclidian algorithm
	// for bezout's identity to set
	// a and b
	new(big.Int).GCD(x, y, m1, m2)

	// fmt.Printf("bezout's: x = %s; y = %s\n", x, y)

	m := new(big.Int).Mul(m1, m2)
	r := new(big.Int).Add(new(big.Int).Mul(new(big.Int).Mul(r1, y), m2), new(big.Int).Mul(new(big.Int).Mul(r2, x), m1))
	r.Mod(r, m)

	// If the residue is less than m bring it positive
	if r.Cmp(big.NewInt(0)) < 0 {
		r.Add(r, m)
	}

	return r, m
}

func startup() {

	changeBinDir()
	limitTime(45) // Ugh generating special RSA keys sucks and I'm not smart enough to make it suck faster

	for i := 0; i < challenge_count; i++ {
		challenge_status[i] = false
	}

	bannerbuf, err := ioutil.ReadFile("./banner.txt")

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read banner: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Fprint(os.Stdout, string(bannerbuf))

	fbuf, err := ioutil.ReadFile("./flag_2022_first.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read flag_2022_first: %s\n", err.Error())
		os.Exit(1)
	}
	flag_2022_first = string(fbuf)

	fbuf, err = ioutil.ReadFile("./flag_2022_second.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read flag_2022_second: %s\n", err.Error())
		os.Exit(1)
	}
	flag_2022_second = string(fbuf)

	fbuf, err = ioutil.ReadFile("./flag_2023.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read flag_2023: %s\n", err.Error())
		os.Exit(1)
	}
	flag_2023 = string(fbuf)

	fbuf, err = ioutil.ReadFile("./sample_code.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read sample_code.txt: %s\n", err.Error())
		os.Exit(1)
	}
	sample_code = string(fbuf)
}


// Change to working directory
func changeBinDir() {
	// read /proc/self/exe
	if dest, err := os.Readlink("/proc/self/exe"); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading link: %s\n", err)
		return
	} else {
		dest = path.Dir(dest)
		if err := os.Chdir(dest); err != nil {
			fmt.Fprintf(os.Stderr, "Error changing directory: %s\n", err)
		}
	}
}


// Limit CPU time to certain number of seconds
func limitTime(secs int) {
	lims := &syscall.Rlimit{
		Cur: uint64(secs),
		Max: uint64(secs),
	}
	if err := syscall.Setrlimit(syscall.RLIMIT_CPU, lims); err != nil {
		if inner_err := syscall.Getrlimit(syscall.RLIMIT_CPU, lims); inner_err != nil {
			fmt.Fprintf(os.Stderr, "Error getting limits: %s\n", inner_err)
		} else {
			if lims.Cur > 0 {
				// A limit was set elsewhere, we'll live with it
				return
			}
		}
		fmt.Fprintf(os.Stderr, "Error setting limits: %s", err)
		os.Exit(-1)
	}
}
