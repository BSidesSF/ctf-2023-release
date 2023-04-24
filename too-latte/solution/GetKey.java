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
