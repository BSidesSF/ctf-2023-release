package org.bsidessf.ctf;

import java.lang.invoke.MethodHandles;


public class Flag
{
  public static void printFlag() {
    int a = MethodHandles.lookup().lookupClass().hashCode();

    System.out.println((new Object() {int t;public String toString() {byte[] buf = new byte[25];t = 1343462807;buf[0] = (byte) ((t-a) >>> 24);t = 2026693101;buf[1] = (byte) ((t-a) >>> 1);t = 2146089908;buf[2] = (byte) ((t+a) >>> 15);t = -483824880;buf[3] = (byte) ((t+a) >>> 21);t = 1174351563;buf[4] = (byte) ((t-a) >>> 12);t = 1609638594;buf[5] = (byte) ((t+a) >>> 21);t = 1087909388;buf[6] = (byte) ((t-a) >>> 5);t = 1432631813;buf[7] = (byte) ((t+a) >>> 24);t = 935671076;buf[8] = (byte) ((t-a) >>> 10);t = 1501451454;buf[9] = (byte) ((t+a) >>> 9);t = 1429601506;buf[10] = (byte) ((t+a) >>> 24);t = 840017620;buf[11] = (byte) ((t-a) >>> 18);t = 569137352;buf[12] = (byte) ((t-a) >>> 2);t = -894683456;buf[13] = (byte) ((t+a) >>> 10);t = 411218479;buf[14] = (byte) ((t-a) >>> 7);t = -35115096;buf[15] = (byte) ((t+a) >>> 1);t = -285779740;buf[16] = (byte) ((t+a) >>> 19);t = 1487744353;buf[17] = (byte) ((t+a) >>> 17);t = -38115808;buf[18] = (byte) ((t-a) >>> 10);t = -434746468;buf[19] = (byte) ((t-a) >>> 18);t = 2115656009;buf[20] = (byte) ((t-a) >>> 6);t = 2052723969;buf[21] = (byte) ((t-a) >>> 10);t = -1062524965;buf[22] = (byte) ((t-a) >>> 16);t = -1412557358;buf[23] = (byte) ((t-a) >>> 10);t = -361286091;buf[24] = (byte) ((t+a) >>> 13);return new String(buf);}}.toString()));

  }
}
