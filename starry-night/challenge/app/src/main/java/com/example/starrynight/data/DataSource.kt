package com.example.starrynight.data
import com.example.starrynight.R
import com.example.starrynight.model.Gallery

class DataSource {
    fun loadGallery(): List<Gallery> {
        return listOf<Gallery>(
            Gallery(R.string.caption1, R.drawable.image1),
            Gallery(R.string.caption2, R.drawable.image2),
            Gallery(R.string.caption3, R.drawable.image3),
            Gallery(R.string.caption4, R.drawable.image4),
            Gallery(R.string.caption5, R.drawable.image5),
            Gallery(R.string.caption6, R.drawable.image6)
        )
    }
}