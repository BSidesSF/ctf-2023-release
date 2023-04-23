## `flat-white` and `flat-white-extra-shot`

Back in February when I worked on
[CVE-2023-0669](https://attackerkb.com/topics/mg883Nbeva/cve-2023-0669/rapid7-analysis),
I had to learn a bunch of Java stuff quickly! The vulnerability is basically a
Java object that's serialized and encrypted with a static key. I'm actually
writing this challenge and CTF write-up on February 8, 2023, only a couple days
after the AttackerKB write-up of the vulnerability. Now that's a pipeline!

In order to reverse engineer the encryption code, I wanted to make sure I was
getting the correct values and found myself trying to call a variety of
functions in their .jar files, some of which were protected or private. I'd
never done that before, so I had to learn how! And it seemed like a useful
skill to pass on to others, hence this challenge.

It turns out, it's super simple. If it's a public function in a .jar file, you
can just call the function from your code:

```java
public class Solve
{
  public static void main(String[] args) {
    org.bsidessf.ctf.Flag.printFlag();
  }
}
```

And then include the .jar file in your classpath when you compile:

```
$ javac -cp '.:FlatWhite.jar' Solve.java
$ java -cp '.:FlatWhite.jar' Solve
CTF{java-java-everywhere}
```

If it's a private function, which is what `flat-white-extra-shot` uses, it's a
bit more complex. Instead of just calling the function, you have to use
reflection:

```java
import java.lang.reflect.Method;

public class Solve
{
  public static void main(String[] args) throws Exception {
    Method method = org.bsidessf.ctf.Flag.class.getDeclaredMethod("printFlag");
    method.setAccessible(true);
    method.invoke(null);
  }
}
```

Then you can compile and run it the same way:

```
$ javac -cp '.:FlatWhiteExtraShot.jar' Solve.java
$ java -cp '.:FlatWhiteExtraShot.jar' Solve
CTF{stronger-java-everywhere}
```

And by the way, the reason for the name, besides being a type of coffee drink
and thematically fitting in with the Java idea, is because my co-creator
[Matir](https://systemoverlord.com/) ordered one at a recent meetup and I
didn't know what it was. So this is to honor him and his great taste in drinks!
