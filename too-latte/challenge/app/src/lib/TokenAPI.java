package org.bsidessf.ctf.toolatte;

import java.io.ByteArrayOutputStream;
import java.io.InputStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.security.Security;
import org.apache.commons.io.IOUtils;
import org.bouncycastle.jcajce.provider.BouncyCastleFipsProvider;

public class TokenAPI {
  public static void initialize() throws Exception {
    Security.addProvider(new BouncyCastleFipsProvider());
    TokenEncryptor.getInstance().initialize();
  }

  public static TokenResponse getValidateResponse(String token) throws Exception {
    TokenAPI.initialize();

    return new TokenResponse(TokenWorker.unbundle(token, getKeyConfig(getVersion(token))));
  }

  public static String getGenerateResponse() throws Exception {
    TokenAPI.initialize();

    return TokenWorker.bundle(Files.readString(Path.of("/var/share/request.xml")), getKeyConfig("2"));
  }

  private static String getVersion(String str) throws Exception {
    int indexOf = str.indexOf(36);
    if (indexOf > -1) {
      return str.substring(indexOf + 1).replace("\r", "").replace("\n", "");
    }
    throw new Exception("Invalid token");
  }

  private static KeyConfig getKeyConfig(String version) throws Exception {
    if(!version.equals("2")) {
      throw new Exception("Invalid version: " + version);
    }

    KeyConfig keyConfig = new KeyConfig();

    InputStream inputStream = null;
    inputStream = TokenAPI.class.getResourceAsStream("keystore-client");

    if(inputStream == null) {
      throw new Exception("Couldn't load embedded keystore!");
    }

    ByteArrayOutputStream byteArrayOutputStream = new ByteArrayOutputStream();
    IOUtils.copy(inputStream, byteArrayOutputStream);
    keyConfig.setKeyStore(byteArrayOutputStream.toByteArray());
    keyConfig.setSigningAlias("client" + version);
    keyConfig.setVerifyingAlias("server" + version);
    keyConfig.setPassword("keystorepw".toCharArray());
    keyConfig.setVersion(version);
    keyConfig.setKeyStoreType("BCFKS");
    if (inputStream != null) {
      IOUtils.closeQuietly(inputStream);
    }
    return keyConfig;
  }
}
