package org.bsidessf.ctf;

import java.lang.invoke.MethodHandles;

public class Flag
{
  private static void printFlag() {
    int a = MethodHandles.lookup().lookupClass().hashCode() ^ new Flag().hashCode();

    System.out.println((new Object() {int t;public String toString() {byte[] buf = new byte[29];t = 281080406;buf[0] = (byte) ((t-a) >>> 7);t = 1736715287;buf[1] = (byte) ((t-a) >>> 20);t = 205173622;buf[2] = (byte) ((t-a) >>> 12);t = 1957310302;buf[3] = (byte) ((t-a) >>> 6);t = 1460936098;buf[4] = (byte) ((t-a) >>> 10);t = -1451624768;buf[5] = (byte) ((t-a) >>> 11);t = -22472705;buf[6] = (byte) ((t-a) >>> 16);t = 1505724756;buf[7] = (byte) ((t+a) >>> 22);t = 1850130167;buf[8] = (byte) ((t-a) >>> 8);t = -1450802081;buf[9] = (byte) ((t-a) >>> 7);t = 1184179022;buf[10] = (byte) ((t+a) >>> 17);t = 115753105;buf[11] = (byte) ((t+a) >>> 14);t = 371458445;buf[12] = (byte) ((t+a) >>> 17);t = 1101294015;buf[13] = (byte) ((t+a) >>> 14);t = -1839869978;buf[14] = (byte) ((t+a) >>> 13);t = -525628450;buf[15] = (byte) ((t-a) >>> 7);t = 2047434053;buf[16] = (byte) ((t-a) >>> 11);t = 823456578;buf[17] = (byte) ((t-a) >>> 4);t = 530297482;buf[18] = (byte) ((t-a) >>> 16);t = -1296211576;buf[19] = (byte) ((t+a) >>> 11);t = 176122434;buf[20] = (byte) ((t+a) >>> 21);t = -1046268445;buf[21] = (byte) ((t-a) >>> 9);t = -1227806326;buf[22] = (byte) ((t+a) >>> 8);t = 1427908414;buf[23] = (byte) ((t+a) >>> 8);t = 1243856245;buf[24] = (byte) ((t-a) >>> 5);t = -768374983;buf[25] = (byte) ((t-a) >>> 1);t = 1508956396;buf[26] = (byte) ((t+a) >>> 12);t = 1962378157;buf[27] = (byte) ((t-a) >>> 23);t = 639951977;buf[28] = (byte) ((t-a) >>> 19);return new String(buf);}}.toString()));

  }
}
