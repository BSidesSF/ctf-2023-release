Generating the keystores:

```
# Create a client store
keytool -v -genkeypair -storetype BCFKS -providerclass org.bouncycastle.jcajce.provider.BouncyCastleFipsProvider -providerpath ../../../../../lib/bc-fips-1.0.2.3.jar -storepass "keystorepw" -keyalg rsa -validity 36500 -keystore ./keystore-client -alias client2 -dname "CN=BSidesSF, OU=CTF, O=BSidesSF, L=San Francisco, ST=CA, C=US"

# Create a server store
keytool -v -genkeypair -storetype BCFKS -providerclass org.bouncycastle.jcajce.provider.BouncyCastleFipsProvider -providerpath ../../../../../lib/bc-fips-1.0.2.3.jar -storepass "keystorepw" -keyalg rsa -validity 36500 -keystore ./keystore-server -alias server2 -dname "CN=BSidesSF, OU=CTF, O=BSidesSF, L=San Francisco, ST=CA, C=US"

# Export a public cert from the server
keytool -v -exportcert -storetype BCFKS -providerclass org.bouncycastle.jcajce.provider.BouncyCastleFipsProvider -providerpath ../../../../../lib/bc-fips-1.0.2.3.jar -storepass "keystorepw" -keystore ./keystore-server -alias server2 -file servercert

# Import the public cert into the client
keytool -v -importcert -storetype BCFKS -providerclass org.bouncycastle.jcajce.provider.BouncyCastleFipsProvider -providerpath ../../../../../lib/bc-fips-1.0.2.3.jar -storepass "keystorepw" -keystore ./keystore-client -alias server2 -file ./servercert

# Validate
keytool -v -list -providerclass org.bouncycastle.jcajce.provider.BouncyCastleFipsProvider -providerpath ../../../../../lib/bc-fips-1.0.2.3.jar -storetype BCFKS -storepass keystorepw -keystore ./keystore-client
keytool -v -list -providerclass org.bouncycastle.jcajce.provider.BouncyCastleFipsProvider -providerpath ../../../../../lib/bc-fips-1.0.2.3.jar -storetype BCFKS -storepass keystorepw -keystore ./keystore-server

# Get rid of the extra files (you don't ever really need the server cert)
rm servercert
mv keystore-server ../../../../../../
```
