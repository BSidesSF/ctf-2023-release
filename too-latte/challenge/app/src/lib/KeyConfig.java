package org.bsidessf.ctf.toolatte;

import java.io.File;
import java.io.FileInputStream;
import java.io.ByteArrayInputStream;
import java.io.FileNotFoundException;
import java.io.InputStream;

public class KeyConfig {
  private File keyStoreFile = null;
  private byte[] keyStore = null;
  private String signingAlias = null;
  private String verifyingAlias = null;
  private char[] password = null;
  private String version = null;
  private String keyStoreType = null;

  public String getSigningAlias() {
    return this.signingAlias;
  }

  public void setSigningAlias(String str) {
    this.signingAlias = str;
  }

  public String getVerifyingAlias() {
    return this.verifyingAlias;
  }

  public void setVerifyingAlias(String str) {
    this.verifyingAlias = str;
  }

  public InputStream getKeyStoreAsStream() throws FileNotFoundException {
    if (this.keyStore != null) {
      return new ByteArrayInputStream(this.keyStore);
    }
    return new FileInputStream(this.keyStoreFile);
  }

  public void setKeyStore(byte[] data) {
    this.keyStore = data;
  }

  public void setKeyStoreFile(File file) {
    this.keyStoreFile = file;
  }

  public char[] getPassword() {
    return this.password;
  }

  public void setPassword(char[] cArr) {
    this.password = cArr;
  }

  public String getVersion() {
    return this.version;
  }

  public void setVersion(String str) {
    this.version = str;
  }

  public String getKeyStoreType() {
    return this.keyStoreType;
  }

  public void setKeyStoreType(String str) {
    this.keyStoreType = str;
  }
}
