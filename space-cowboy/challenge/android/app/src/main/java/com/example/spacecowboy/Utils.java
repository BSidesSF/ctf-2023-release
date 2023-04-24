package com.example.spacecowboy;

import android.util.Log;

import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.time.LocalDate;
import java.time.ZoneId;

public class Utils {
    private static String formCouponPrefix(String uid){
        String part1 = uid.substring(0, 3);
        LocalDate date = LocalDate.now(ZoneId.of("Etc/UTC"));
        int temp = date.getDayOfMonth();
        String part2 = String.format("%02d", temp);
        String part3 = date.getMonth().toString();
        part3 = part3.substring(0,2);
        String result = part1 + part2 + part3;
        //Log.d("Raw coupon value:", result);
        return result.toUpperCase();
    }
    private static byte[] couponDigest(String couponPrefix) {
        MessageDigest md = null;
        try {
            md = MessageDigest.getInstance("MD5");
        } catch (NoSuchAlgorithmException e) {
            throw new RuntimeException(e);
        }
        md.update(couponPrefix.getBytes());
        return md.digest();
    }

    private static String couponHex(byte[] bytes) {
        StringBuilder sb = new StringBuilder();
        for (byte b : bytes) {
            sb.append(String.format("%02x", b));
        }
        return sb.toString();
    }

    public static boolean validateCoupon(String uid, String coupon){
        if (coupon.length() != 10){
            return false;
        }
        coupon = coupon.substring(0,8).toUpperCase();
        byte[] couponBytes = couponDigest(formCouponPrefix(uid));
        String couponCode = couponHex(couponBytes);
        couponCode = couponCode.substring(0,8).toUpperCase();
        //Log.d("Coupon prefix:",couponCode);
        return coupon.equals(couponCode);
    }
}
