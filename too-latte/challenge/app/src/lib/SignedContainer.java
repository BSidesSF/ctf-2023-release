package org.bsidessf.ctf.toolatte;

import java.io.Serializable;

public class SignedContainer implements Serializable {
  private byte[] data = null;

  public byte[] getData() {
    return this.data;
  }

  public void setData(byte[] data) {
    this.data = data;
  }
}
