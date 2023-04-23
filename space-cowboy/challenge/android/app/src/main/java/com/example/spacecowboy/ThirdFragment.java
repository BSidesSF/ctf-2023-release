package com.example.spacecowboy;



import android.os.Bundle;
import android.util.Log;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.Button;
import android.widget.TextView;
import androidx.annotation.NonNull;
import androidx.fragment.app.Fragment;
import androidx.navigation.fragment.NavHostFragment;
import com.example.spacecowboy.databinding.FragmentThirdBinding;
import com.google.android.gms.tasks.OnCompleteListener;
import com.google.android.gms.tasks.Task;
import com.google.firebase.auth.FirebaseAuth;
import com.google.firebase.auth.FirebaseUser;
import com.google.firebase.auth.GetTokenResult;

import org.jetbrains.annotations.NotNull;

import java.io.IOException;
import java.util.concurrent.TimeUnit;

import okhttp3.Call;
import okhttp3.Callback;
import okhttp3.FormBody;
import okhttp3.HttpUrl;
import okhttp3.MediaType;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.RequestBody;
import okhttp3.Response;

public class ThirdFragment extends Fragment {

    private FragmentThirdBinding binding;
    final OkHttpClient client = new OkHttpClient().newBuilder()
            .connectTimeout(2, TimeUnit.MINUTES)
            .readTimeout(2, TimeUnit.MINUTES)
            .writeTimeout(2, TimeUnit.MINUTES)
            .build();
    private String responseStr = null;
    private TextView flagTextView;


    @Override
    public View onCreateView(
            LayoutInflater inflater, ViewGroup container,
            Bundle savedInstanceState
    ) {
        binding = FragmentThirdBinding.inflate(inflater, container, false);
        return binding.getRoot();
    }

    public void onViewCreated(@NonNull View view, Bundle savedInstanceState) {
        super.onViewCreated(view, savedInstanceState);
        // Get the flag text view, will update on response
        flagTextView = (TextView)getView().findViewById(R.id.flagTextView);
        // Get a valid Firebase user token
        getidToken();
        // Handle the return to home button click
        binding.buttonThird.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                NavHostFragment.findNavController(ThirdFragment.this)
                        .navigate(R.id.action_ThirdFragment_to_FirstFragment);

        }
    });
        // Handle the get flag button
        Button getFlag = (Button)getView().findViewById(R.id.flagFetchButton);
        getFlag.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                getidToken();
            }
        });
    }

    @Override
    public void onDestroyView() {
        super.onDestroyView();
        binding = null;
    }
    // Helper functions
    // Get a valid Firebase user token
    private void getidToken(){
        FirebaseUser mUser = FirebaseAuth.getInstance().getCurrentUser();
        mUser.getIdToken(true)
                .addOnCompleteListener(new OnCompleteListener<GetTokenResult>() {
                    public void onComplete(@NonNull Task<GetTokenResult> task) {
                        if (task.isSuccessful()) {
                            String idToken = task.getResult().getToken();
                            //Log.d("token",idToken);
                            getFlag(idToken);
                        } else {
                            Log.w("Token Fetch","Failed");
                        }
                    }
                });
    }

    // Request the flag from the server
    // Server will give the flag if user has 500 coins
    private void getFlag(String token){
        int coins = ((MainActivity) getActivity()).getScore();
        // todo update server URL
        String BASE_URL = "https://space-cowboy-8a2ef95e.challenges.bsidessf.net";
        //String BASE_URL = "http://10.0.2.2:8000";
        if (coins >= 500) {

            RequestBody formBody = new FormBody.Builder()
                    .add("token", token)
                    .build();
            Request request = new Request.Builder()
                    .url(BASE_URL + "/get-flag")
                    .post(formBody)
                    .build();
            client.newCall(request).enqueue(new Callback() {
                @Override
                public void onFailure(@NotNull Call call, @NotNull IOException e) {
                    e.printStackTrace();
                }

                @Override
                public void onResponse(@NotNull Call call, @NotNull Response response) throws IOException {
                        if (response.isSuccessful()) {
                            responseStr = response.body().string();
                            Log.d("Response:",responseStr);
                            if (response.code() != 200){
                                responseStr = "Error fetching flag";
                            }
                            getActivity().runOnUiThread(new Runnable() {
                                @Override
                                public void run() {
                                    // Update the flag with the response string
                                    flagTextView.setText(responseStr);

                                }
                            });

                            Log.d("Response:", responseStr);
                        }
                }
            });

        }
        else{
            getActivity().runOnUiThread(new Runnable() {
                @Override
                public void run() {
                    // Update the flag with the response string
                    flagTextView.setText("You need 500 coins to get the flag!");

                }
            });
        }
    }


}