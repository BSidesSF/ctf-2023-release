package org.bsidessf.ctf.toolatte;

import java.io.Serializable;

public final class TokenResponse implements Serializable {
  private String response;

  public TokenResponse(String response) {
    this.response = response;
  }

  public String getResponse() {
    return this.response;
  }
}
