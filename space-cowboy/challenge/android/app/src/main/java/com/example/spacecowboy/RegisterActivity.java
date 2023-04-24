package com.example.spacecowboy;

import androidx.annotation.RequiresApi;
import androidx.appcompat.app.AppCompatActivity;

import android.content.Intent;
import android.os.Build;
import android.security.keystore.KeyGenParameterSpec;
import android.security.keystore.KeyProperties;
import android.text.TextUtils;
import android.util.Base64;
import android.util.Patterns;
import android.widget.EditText;
import android.widget.Button;
import android.widget.Toast;
import android.os.Bundle;
import android.view.View;
import android.util.Log;

import androidx.annotation.NonNull;


import com.google.android.gms.tasks.OnCompleteListener;
import com.google.android.gms.tasks.OnFailureListener;
import com.google.android.gms.tasks.OnSuccessListener;
import com.google.android.gms.tasks.Task;
import com.google.common.base.Splitter;
import com.google.firebase.auth.AuthResult;
import com.google.firebase.auth.FirebaseAuth;
import com.google.firebase.firestore.FirebaseFirestore;
import java.io.IOException;
import java.util.HashMap;
import java.util.Map;


public class RegisterActivity extends AppCompatActivity {
    private FirebaseAuth mAuth;
    private static final String TAG = "EmailPassword";
    EditText mEmail;
    EditText mPassword;
    EditText mPassword2;
    Button mButton;
    private FirebaseFirestore mDB;
    private static String dbPath = "users";
    private static String uid = null;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_register);
        // For Firebase Auth
        mAuth = FirebaseAuth.getInstance();
        // For Firebase Database
        mDB = FirebaseFirestore.getInstance();
        // Getting UI elements
        mEmail = (EditText) findViewById(R.id.username);
        mPassword = (EditText) findViewById(R.id.password);
        mPassword2 = (EditText) findViewById(R.id.password2);
        mButton = (Button) findViewById(R.id.register);
        mButton.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                createAccount(mEmail.getText().toString(), mPassword.getText().toString());
            }
        });
    }

    private void createAccount(String email, String password) {
        Log.d(TAG, "createAccount:" + email);
        if (!validateForm()) {
            return;
        }


        // [START create_user_with_email]
        mAuth.createUserWithEmailAndPassword(email, password)
                .addOnCompleteListener(this, new OnCompleteListener<AuthResult>() {
                    @RequiresApi(api = Build.VERSION_CODES.O)
                    @Override
                    public void onComplete(@NonNull Task<AuthResult> task) {
                        if (task.isSuccessful()) {
                            // Sign in success, update UI with the signed-in user's information
                            Log.d(TAG, "createUserWithEmail:success");
                            uid = mAuth.getCurrentUser().getUid();
                            // Store the key
                            storeDb("100");
                            // Start Main Activity
                            Intent intent = new Intent(RegisterActivity.this, MainActivity.class);
                            startActivity(intent);
                        } else {
                            // If sign in fails, display a message to the user.
                            Log.w(TAG, "createUserWithEmail:failure", task.getException());
                            Toast.makeText(RegisterActivity.this, "Authentication failed.",
                                    Toast.LENGTH_SHORT).show();
                        }
                    }
                });
        // [END create_user_with_email]
    }

    // Make sure email and password is entered
    // Make sure password and confirm password match

    private boolean validateForm() {
        boolean valid = true;
        String email = mEmail.getText().toString();
        CharSequence emailChars = mEmail.getText();
        if (TextUtils.isEmpty(email)) {
            mEmail.setError("Required.");
            valid = false;
        } else if (!validateEmail(emailChars)) {
            mEmail.setError("Should be valid Email");
            valid = false;
        } else {
            mEmail.setError(null);
        }

        String password = mPassword.getText().toString();
        if (TextUtils.isEmpty(password)) {
            mPassword.setError("Required.");
            valid = false;
        } else {
            mPassword.setError(null);
        }

        if (password.length() < 6) {
            mPassword.setError("Password must be atleast 6 characters");
            valid = false;
        }

        String password2 = mPassword2.getText().toString();
        if (TextUtils.isEmpty(password)) {
            mPassword2.setError("Required.");
            valid = false;
        } else {
            mPassword2.setError(null);
        }

        if (!password.equals(password2)) {
            mPassword2.setError("Passwords should match.");
            valid = false;
        }
        return valid;
    }


    // Make sure user entered an email address
    private boolean validateEmail(CharSequence input) {
        return Patterns.EMAIL_ADDRESS.matcher(input).matches();
    }

    //Iniitialize user's coins to 100
    @RequiresApi(api = Build.VERSION_CODES.O)
    private void storeDb(String coins) {
        Map<String, Object> user = new HashMap<>();
        user.put("coins", coins);
        mDB.collection(dbPath).document(uid)
                .set(user)
                .addOnSuccessListener(new OnSuccessListener<Void>() {
                    @Override
                    public void onSuccess(Void aVoid) {
                        Log.d("Writing score to DB:", "success");
                    }
                })
                .addOnFailureListener(new OnFailureListener() {
                    @Override
                    public void onFailure(@NonNull Exception e) {
                        Log.w("Writing score to DB:", "error", e);
                    }
                });

    }

}