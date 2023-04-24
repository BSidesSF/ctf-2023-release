package org.bsidessf.ctf.toolatte;

public interface EncryptionEngine {
    void init() throws Exception;

    byte[] encrypt(byte[] data) throws Exception;

    byte[] decrypt(byte[] data) throws Exception;
}
