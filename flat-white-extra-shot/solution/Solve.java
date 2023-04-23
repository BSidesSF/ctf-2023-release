import java.lang.reflect.Method;

public class Solve
{
  public static void main(String[] args) throws Exception {
    Method method = org.bsidessf.ctf.Flag.class.getDeclaredMethod("printFlag");
    method.setAccessible(true);
    method.invoke(null);
  }
}
