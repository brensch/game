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
        Seq.setContext(getApplicationContext());
        view = new EbitenView(this);
        setContentView(view);
    }

    @Override
    protected void onPause() {
        super.onPause();
        if (view != null) {
            view.suspendGame();
        }
    }

    @Override
    protected void onResume() {
        super.onResume();
        if (view != null) {
            view.resumeGame();
        }
    }
}