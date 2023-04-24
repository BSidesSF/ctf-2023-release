package org.bsidessf.ctf.toolatte;

import java.io.UnsupportedEncodingException;

public class Encryptor {
    private static String CHARSET = "UTF-16LE";
    private EncryptionEngine engine;
    private String alias = null;

    public Encryptor(EncryptionEngine encryptionEngine) {
        this.engine = null;
        this.engine = encryptionEngine;
    }

    public void setAlias(String str) {
        this.alias = str;
    }

    public String getAlias() {
        return this.alias;
    }

    public byte[] encryptFromString(String str) throws Exception {
        try {
            return this.engine.encrypt(str.getBytes(CHARSET));
        } catch (Throwable th) {
            throw new Exception(th.getMessage(), th);
        }
    }

    public byte[] encryptFromBytes(byte[] data) throws Exception {
        return this.engine.encrypt(data);
    }

    public String decryptToString(byte[] data) throws Exception {
        try {
            return new String(this.engine.decrypt(data), CHARSET);
        } catch (UnsupportedEncodingException e) {
            throw new Exception(e.getMessage(), e);
        }
    }

    public byte[] decryptToBytes(byte[] data) throws Exception {
        return this.engine.decrypt(data);
    }

    public byte[] getKeyBytes() {
        if (this.engine instanceof StandardEncryptionEngine) {
            return ((StandardEncryptionEngine) this.engine).getKeyBytes();
        }
        return null;
    }

    public boolean isCustom() {
        return !(this.engine instanceof StandardEncryptionEngine);
    }
}
