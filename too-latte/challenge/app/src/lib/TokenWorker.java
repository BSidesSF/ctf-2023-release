package org.bsidessf.ctf.toolatte;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.ObjectInputStream;
import java.io.ObjectOutputStream;
import java.io.UnsupportedEncodingException;
import java.nio.charset.StandardCharsets;
import java.security.InvalidKeyException;
import java.security.Key;
import java.security.KeyStore;
import java.security.KeyStoreException;
import java.security.NoSuchAlgorithmException;
import java.security.NoSuchProviderException;
import java.security.PrivateKey;
import java.security.PublicKey;
import java.security.Signature;
import java.security.SignatureException;
import java.security.SignedObject;
import java.security.UnrecoverableKeyException;
import java.security.cert.Certificate;
import java.security.cert.CertificateException;
import java.util.zip.GZIPInputStream;
import java.util.zip.GZIPOutputStream;
import org.apache.commons.codec.binary.Base64;
import org.apache.commons.io.IOUtils;

public class TokenWorker {
  public static String unbundle(String token, KeyConfig keyConfig) throws Exception {
    token = token.substring(0, token.indexOf("$"));

    return new String(decompress(verify(decrypt(decode(token.getBytes(StandardCharsets.UTF_8)), keyConfig.getVersion()), keyConfig)), StandardCharsets.UTF_8);
  }

  public static String bundle(String data, KeyConfig keyConfig) throws Exception {
    return encode(encrypt(sign(compress(data.getBytes(StandardCharsets.UTF_8)), keyConfig), keyConfig.getVersion())) + "$2";
  }

  private static String encode(byte[] data) throws UnsupportedEncodingException {
    return new String(Base64.encodeBase64(data, false, true), StandardCharsets.UTF_8);
  }

  private static byte[] decode(byte[] data) {
    return Base64.decodeBase64(data);
  }

  private static byte[] encrypt(byte[] data, String version) throws Exception {
    return TokenEncryptor.getInstance().encrypt(data);
  }

  private static byte[] decrypt(byte[] data, String version) throws Exception {
    return TokenEncryptor.getInstance().decrypt(data);
  }

  private static byte[] verify(byte[] data, KeyConfig keyConfig) throws Exception {
    ObjectInputStream objectInputStream = null;
    PublicKey publicKey = getPublicKey(keyConfig);
    ObjectInputStream objectInputStream2 = new ObjectInputStream(new ByteArrayInputStream(data));

    SignedObject signedObject = (SignedObject) objectInputStream2.readObject();

    if (!signedObject.verify(publicKey, Signature.getInstance("SHA512withRSA"))) {
        throw new IOException("Unable to verify signature! Did you send us a Token Request by mistake?");
    }

    byte[] outData = ((SignedContainer) signedObject.getObject()).getData();
    if (objectInputStream2 != null) {
        objectInputStream2.close();
    }
    return outData;
  }

  private static PublicKey getPublicKey(KeyConfig keyConfig) throws Exception {
    InputStream inputStream = null;
    KeyStore keyStore = KeyStore.getInstance(keyConfig.getKeyStoreType());
    InputStream keyStoreAsStream = keyConfig.getKeyStoreAsStream();
    keyStore.load(keyStoreAsStream, keyConfig.getPassword());
    Certificate certificate = keyStore.getCertificate(keyConfig.getVerifyingAlias());
    if (certificate == null) {
      throw new KeyStoreException("Specified public key not found: " + keyConfig.getVerifyingAlias());
    }
    PublicKey publicKey = certificate.getPublicKey();
    if (keyStoreAsStream != null) {
      keyStoreAsStream.close();
    }
    return publicKey;
  }

  private static byte[] sign(byte[] data, KeyConfig keyConfig) throws Exception {
    Signature signature = Signature.getInstance("SHA512withRSA");

    ObjectOutputStream objectOutputStream = null;
    PrivateKey privateKey = getPrivateKey(keyConfig);
    SignedContainer signedContainer = new SignedContainer();
    signedContainer.setData(data);
    SignedObject signedObject = new SignedObject(signedContainer, privateKey, signature);
    ByteArrayOutputStream byteArrayOutputStream = new ByteArrayOutputStream();
    objectOutputStream = new ObjectOutputStream(byteArrayOutputStream);
    objectOutputStream.writeObject(signedObject);
    byte[] byteArray = byteArrayOutputStream.toByteArray();
    if (objectOutputStream != null) {
        objectOutputStream.close();
    }
    return byteArray;
  }

  private static byte []decompress(byte []data) throws Exception {
    GZIPInputStream gZIPInputStream = new GZIPInputStream(new ByteArrayInputStream(data));
    ByteArrayOutputStream byteArrayOutputStream = new ByteArrayOutputStream();
    IOUtils.copy(gZIPInputStream, byteArrayOutputStream);

    return byteArrayOutputStream.toByteArray();
  }

  private static byte []compress(byte []data) throws Exception {
    ByteArrayOutputStream byteArrayOutputStream = new ByteArrayOutputStream();
    GZIPOutputStream gZIPOutputStream = null;

    gZIPOutputStream = new GZIPOutputStream(byteArrayOutputStream);
    gZIPOutputStream.write(data);
    if (gZIPOutputStream != null) {
      gZIPOutputStream.close();
    }
    return byteArrayOutputStream.toByteArray();
  }

  private static PrivateKey getPrivateKey(KeyConfig keyConfig) throws Exception {
    InputStream inputStream = null;
    KeyStore keyStore = KeyStore.getInstance(keyConfig.getKeyStoreType());
    InputStream keyStoreAsStream = keyConfig.getKeyStoreAsStream();
    keyStore.load(keyStoreAsStream, keyConfig.getPassword());
    Key key = keyStore.getKey(keyConfig.getSigningAlias(), keyConfig.getPassword());
    if (key == null || !(key instanceof PrivateKey)) {
      throw new KeyStoreException("Specified key not found: " + keyConfig.getSigningAlias());
    }
    PrivateKey privateKey = (PrivateKey) key;
    if (keyStoreAsStream != null) {
      keyStoreAsStream.close();
    }
    return privateKey;
  }
}
