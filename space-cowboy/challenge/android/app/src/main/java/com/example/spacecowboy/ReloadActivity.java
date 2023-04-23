package com.example.spacecowboy;

import android.app.Activity;
import android.os.Bundle;
import android.util.Log;
import android.view.View;
import android.widget.Button;
import android.widget.EditText;
import android.widget.Toast;

import androidx.annotation.NonNull;
import androidx.navigation.fragment.NavHostFragment;
import androidx.navigation.ui.AppBarConfiguration;
import com.example.spacecowboy.databinding.ActivityMainBinding;
import com.example.spacecowboy.databinding.ActivityReloadBinding;
import com.example.spacecowboy.databinding.FragmentFirstBinding;
import com.example.spacecowboy.models.User;
import com.google.android.gms.tasks.OnCompleteListener;
import com.google.android.gms.tasks.OnFailureListener;
import com.google.android.gms.tasks.OnSuccessListener;
import com.google.android.gms.tasks.Task;
import com.google.firebase.auth.FirebaseAuth;
import com.google.firebase.auth.FirebaseUser;
import com.google.firebase.auth.GetTokenResult;
import com.google.firebase.firestore.FirebaseFirestore;

import org.jetbrains.annotations.NotNull;

import java.io.IOException;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.TimeUnit;

import okhttp3.Call;
import okhttp3.Callback;
import okhttp3.FormBody;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.RequestBody;
import okhttp3.Response;

public class ReloadActivity extends Activity {
    private FirebaseAuth mAuth;
    private FirebaseFirestore mDB;
    private AppBarConfiguration appBarConfiguration;
    private ActivityReloadBinding binding;
    private String responseStr;
    final OkHttpClient client = new OkHttpClient().newBuilder()
            .connectTimeout(2, TimeUnit.MINUTES)
            .readTimeout(2, TimeUnit.MINUTES)
            .writeTimeout(2, TimeUnit.MINUTES)
            .build();
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_reload);
        // Get Firebase Auth instance to get userid
        mAuth = FirebaseAuth.getInstance();
        String uid = mAuth.getCurrentUser().getUid();
        Button submitButton = (Button)findViewById(R.id.reloadButton);
        submitButton.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view){
                EditText editText = (EditText)findViewById(R.id.couponEditText);
                String coupon = editText.getText().toString();
                if(Utils.validateCoupon(uid,coupon)){
                    // If coupon prefix is valid, send to server for redemption
                    getidToken(coupon);
                }
                else{
                    Toast.makeText(ReloadActivity.this,R.string.invalid_coupon, Toast.LENGTH_LONG).show();
                }
            }
        });

    }
    // Get a valid Firebase token for current user
    private void getidToken(String coupon){
        FirebaseUser mUser = FirebaseAuth.getInstance().getCurrentUser();
        mUser.getIdToken(true)
                .addOnCompleteListener(new OnCompleteListener<GetTokenResult>() {
                    public void onComplete(@NonNull Task<GetTokenResult> task) {
                        if (task.isSuccessful()) {
                            String idToken = task.getResult().getToken();
                            //Log.d("token",idToken);
                            redeemCoupon(idToken, coupon);
                        } else {
                            Log.w("Token Fetch","Failed");
                        }
                    }
                });
    }
    // Redeem the coupon
    private void redeemCoupon(String token, String coupon){
        // todo update server URL
        //String BASE_URL = "http://10.0.2.2:8000";
        String BASE_URL = "https://space-cowboy-8a2ef95e.challenges.bsidessf.net";
        RequestBody formBody = new FormBody.Builder()
                    .add("token", token)
                    .add("coupon",coupon)
                    .build();
            Request request = new Request.Builder()
                    .url(BASE_URL + "/redeem-coupon")
                    .post(formBody)
                    .build();
            //Log.d("URL",request.url().toString());
            client.newCall(request).enqueue(new Callback() {
                @Override
                public void onFailure(@NotNull Call call, @NotNull IOException e) {
                    e.printStackTrace();
                }

                @Override
                public void onResponse(@NotNull Call call, @NotNull Response response) throws IOException {
                    responseStr = response.body().string();
                    int code = response.code();
                    Activity activity = ReloadActivity.this;
                    if (code == 200){
                        activity.runOnUiThread(new Runnable() {
                            public void run() {
                                Toast.makeText(activity,R.string.coupon_success, Toast.LENGTH_LONG).show();
                            }
                        });
                    }
                    else {
                        activity.runOnUiThread(new Runnable() {
                            public void run() {
                                Toast.makeText(activity,R.string.coupon_fail, Toast.LENGTH_LONG).show();
                            }
                        });
                    }
                    Log.d("Response:",responseStr);
                }
            });
    }

    /*private void storeDb(String coupon){
        FirebaseAuth mAuth = FirebaseAuth.getInstance();
        String uid = mAuth.getCurrentUser().getUid();
        Map<String, Object> couponMap = new HashMap<>();
        couponMap.put("redeemed","true" );
        mDB = FirebaseFirestore.getInstance();
        mDB.collection("users").document(uid).collection("coupons").document(coupon)
                .set(couponMap)
                .addOnSuccessListener(new OnSuccessListener<Void>() {
                    @Override
                    public void onSuccess(Void aVoid) {
                        Log.d("Writing coupon to DB:", "success");
                    }

                })
                .addOnFailureListener(new OnFailureListener() {
                    @Override
                    public void onFailure(@NonNull Exception e) {
                        Log.w("Writing coupon to DB:", "error", e);

                    }
                });

    }*/
}