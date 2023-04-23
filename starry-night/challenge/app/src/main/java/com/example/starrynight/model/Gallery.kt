package com.example.starrynight.model
import androidx.annotation.DrawableRes
import androidx.annotation.StringRes
data class Gallery(
    @StringRes val stringResourceId: Int,
    @DrawableRes val imageResourceId: Int
)
