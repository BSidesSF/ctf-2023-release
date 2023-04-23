package com.example.spacecowboy;

import static android.content.ContentValues.TAG;

import android.content.Intent;
import android.os.Bundle;

import com.example.spacecowboy.models.User;
import androidx.annotation.Nullable;
import androidx.appcompat.app.AppCompatActivity;

import android.util.Log;
import android.view.View;

import androidx.databinding.BindingAdapter;
import androidx.lifecycle.ViewModelProvider;
import androidx.navigation.NavController;
import androidx.navigation.Navigation;
import androidx.navigation.ui.AppBarConfiguration;
import androidx.navigation.ui.NavigationUI;

import com.example.spacecowboy.databinding.ActivityMainBinding;
import androidx.databinding.DataBindingUtil;
import com.google.firebase.auth.FirebaseAuth;
import com.google.firebase.firestore.DocumentReference;
import com.google.firebase.firestore.DocumentSnapshot;
import com.google.firebase.firestore.EventListener;
import com.google.firebase.firestore.FirebaseFirestore;
import com.google.firebase.firestore.FirebaseFirestoreException;
import android.view.Menu;
import android.view.MenuItem;
import android.widget.TextView;

import java.util.Map;

public class MainActivity extends AppCompatActivity {

    private AppBarConfiguration appBarConfiguration;
    private ActivityMainBinding binding;
    private FirebaseAuth mAuth;
    private FirebaseFirestore mDB;
    private User user = null;
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        // Using data binding to update the score
        binding = DataBindingUtil.setContentView(this, R.layout.activity_main);
        binding.setLifecycleOwner(this);
        // For Firebase Auth
        mAuth = FirebaseAuth.getInstance();
        user = new User(mAuth.getCurrentUser().getUid());
        // Attach the user to the binding
        binding.setUser(user);
        // Handling the toolbar and navigation to fragments
        setSupportActionBar(binding.toolbar);
        NavController navController = Navigation.findNavController(this, R.id.nav_host_fragment_content_main);
        appBarConfiguration = new AppBarConfiguration.Builder(navController.getGraph()).build();
        NavigationUI.setupActionBarWithNavController(this, navController, appBarConfiguration);
        // For Firebase Database
        mDB = FirebaseFirestore.getInstance();
        // Handler that will update the score anytime value changes on server side
        updateScore();
        // Reloading coins
        binding.fab.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                startActivity(new Intent(MainActivity.this, ReloadActivity.class));
            }
        });
    }
    @Override
    public boolean onCreateOptionsMenu(Menu menu) {
        // Inflate the menu; this adds items to the action bar if it is present.
        getMenuInflater().inflate(R.menu.menu_main, menu);
        return true;
    }

    @Override
    public boolean onOptionsItemSelected(MenuItem item) {
        // Handle action bar item clicks here. The action bar will
        // automatically handle clicks on the Home/Up button, so long
        // as you specify a parent activity in AndroidManifest.xml.
        int id = item.getItemId();
        return super.onOptionsItemSelected(item);
    }

    @Override
    public boolean onSupportNavigateUp() {
        NavController navController = Navigation.findNavController(this, R.id.nav_host_fragment_content_main);
        return NavigationUI.navigateUp(navController, appBarConfiguration)
                || super.onSupportNavigateUp();
    }
    // Helper functions
    // Snapshot listener will listen for updates on Firestore and update score in App
    public void updateScore(){
        final DocumentReference docRef = mDB.collection("users").document(user.id);
        docRef.addSnapshotListener(new EventListener<DocumentSnapshot>() {
            @Override
            public void onEvent(@Nullable DocumentSnapshot snapshot,
                                @Nullable FirebaseFirestoreException e) {
                if (e != null) {
                    Log.w(TAG, "Listen failed.", e);
                    return;
                }

                if (snapshot != null && snapshot.exists()) {
                    Map<String, Object> temp = snapshot.getData();
                    Log.d(TAG, "Current data: " + temp);
                    user.coins.set((String)temp.get("coins"));
                } else {
                    Log.d(TAG, "Current data: null");
                }
            }
        });
    }
    // Fetch the current score, data binding should ensure user variable and firestore are in sync
    public int getScore(){
        return Integer.parseInt(user.coins.get());

    }



}