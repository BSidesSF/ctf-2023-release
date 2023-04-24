package org.bsidessf.ctf.toolatte;

import javax.crypto.Cipher;
import javax.crypto.spec.IvParameterSpec;
import javax.crypto.spec.SecretKeySpec;

public class StandardEncryptionEngine implements EncryptionEngine {
    private Cipher encryptionCipher;
    private Cipher decryptionCipher;
    private byte[] keyBytes;

    public StandardEncryptionEngine(byte[] key, byte[] iv, String algorithm, String padding) throws Exception {
        this.encryptionCipher = null;
        this.decryptionCipher = null;
        this.keyBytes = null;
        try {
            SecretKeySpec secretKeySpec = new SecretKeySpec(key, algorithm);
            this.encryptionCipher = Cipher.getInstance(padding);
            this.encryptionCipher.init(1, secretKeySpec, new IvParameterSpec(iv));
            this.decryptionCipher = Cipher.getInstance(padding);
            this.decryptionCipher.init(2, secretKeySpec, new IvParameterSpec(iv));
            this.keyBytes = key;
        } catch (Throwable th) {
            throw new Exception("Error initializing symmetric key", th);
        }
    }

    public StandardEncryptionEngine(Cipher cipher, Cipher cipher2) {
        this.encryptionCipher = null;
        this.decryptionCipher = null;
        this.keyBytes = null;
        this.encryptionCipher = cipher;
        this.decryptionCipher = cipher2;
    }

    @Override // com.linoma.security.commons.crypto.cipher.EncryptionEngine
    public byte[] encrypt(byte[] data) throws Exception {
        byte[] doFinal;
        synchronized (this.encryptionCipher) {
            doFinal = this.encryptionCipher.doFinal(data);
        }
        return doFinal;
    }

    @Override // com.linoma.security.commons.crypto.cipher.EncryptionEngine
    public byte[] decrypt(byte[] data) throws Exception {
        byte[] doFinal;
        synchronized (this.decryptionCipher) {
            doFinal = this.decryptionCipher.doFinal(data);
        }
        return doFinal;
    }

    /* JADX INFO: Access modifiers changed from: protected */
    public byte[] getKeyBytes() {
        return this.keyBytes;
    }

    @Override // com.linoma.security.commons.crypto.cipher.EncryptionEngine
    public void init() {
    }
}
