package org.bsidessf.ctf.toolatte;

import javax.crypto.SecretKeyFactory;
import javax.crypto.spec.PBEKeySpec;

public class TokenEncryptor {
    public static final String VERSION_1 = "1";
    public static final String VERSION_2 = "2";
    private static final byte[] IV = {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16};
    private static final String KEY_ALGORITHM = "AES";
    private static final String CIPHER_ALGORITHM = "AES/CBC/PKCS5Padding";
    private boolean initialized = false;
    private Encryptor encryptor = null;
    private static final TokenEncryptor INSTANCE = new TokenEncryptor();

    private TokenEncryptor() {
    }

    public void initialize() throws Exception {
      this.encryptor = new Encryptor(new StandardEncryptionEngine(TokenEncryptor.getInitializationValue(), IV, "AES", "AES/CBC/PKCS5Padding"));
      this.initialized = true;
    }

    public static TokenEncryptor getInstance() {
      return TokenEncryptor.INSTANCE;
    }

    public byte[] encrypt(byte[] data) throws Exception {
      if (!this.initialized) {
        throw new IllegalStateException("The AESEncryptor has not been initialized");
      }

      return this.encryptor.encryptFromBytes(data);
    }

    public byte[] decrypt(byte[] data) throws Exception {
      if (!this.initialized) {
        throw new IllegalStateException("The Token Encryptor has not been initialized");
      }

      return this.encryptor.decryptToBytes(data);
    }

    private static byte[] getInitializationValue() throws Exception {
      return SecretKeyFactory.getInstance("PBKDF2WithHmacSHA1").generateSecret(new PBEKeySpec(new String("cafelatteTokenP@$$wrd".getBytes(), "UTF-8").toCharArray(), new byte[]{12, 56, 72, 86, 73, 99, 35, 44, 35, 97, 45, 45, 89, 23, 33, 67}, 3392, 256)).getEncoded();
    }
}
