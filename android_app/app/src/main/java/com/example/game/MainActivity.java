package com.example.game;

import androidx.appcompat.app.AppCompatActivity;
import android.os.Bundle;
import go.Seq;
import com.example.game.yourgamemobile.EbitenView;

public class MainActivity extends AppCompatActivity {
    private EbitenView view;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        Seq.setContext(this);
        view = new EbitenView(this);
        setContentView(view);
    }
}