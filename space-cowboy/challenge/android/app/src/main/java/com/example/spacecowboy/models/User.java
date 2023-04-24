package com.example.spacecowboy.models;
import androidx.annotation.Keep;
import com.google.firebase.firestore.IgnoreExtraProperties;

import androidx.databinding.BaseObservable;
import androidx.databinding.Bindable;
import androidx.databinding.ObservableField;
import androidx.databinding.ObservableInt;
@Keep
@IgnoreExtraProperties
public class User extends BaseObservable {
    public String id;
    // Needs to be public to use built-in getter/setter
    @Bindable
    public ObservableField<String> coins = new ObservableField<>();
    public User() {
        // Default constructor required for calls to DataSnapshot.getValue(Post.class)
    }
    public User(String id) {
        this.id = id;
        this.coins.set("0");
    }
    public User(String id, String coins) {
        this.id = id;
        this.coins.set(coins);
    }
}