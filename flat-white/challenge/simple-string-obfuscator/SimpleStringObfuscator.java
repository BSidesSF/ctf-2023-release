import java.io.PrintStream;
import java.util.Random;

public class SimpleStringObfuscator {
    public static void main(String[] args) {
        if (args != null && args.length > 0) {
            Random r = new Random(System.currentTimeMillis());
            byte[] b = args[0].getBytes();
            int c = b.length;
            PrintStream o = System.out;

            o.print("(new Object() {");
            o.print("int t;");
            o.print("public String toString() {");
            o.print("byte[] buf = new byte[");
            o.print(c);
            o.print("];");

            for (int i = 0; i < c; ++i) {
                int t = r.nextInt();
                int f = r.nextInt(24) + 1;
                boolean plus = (r.nextInt() % 2) == 0;

                t = (t & ~(0xff << f)) | (b[i] << f);

                o.print("t = ");
                if(plus) {
                  o.print(t + 205029188);
                } else {
                  o.print(t - 205029188);
                }
                o.print(";");
                o.print("buf[");
                o.print(i);
                if(plus) {
                  o.print("] = (byte) ((t-a) >>> ");
                } else {
                  o.print("] = (byte) ((t+a) >>> ");
                }
                o.print(f);
                o.print(");");
            }

            o.print("return new String(buf);");
            o.print("}}.toString())");
            o.println();
        }
    }
}
