package com.example.game;

import androidx.appcompat.app.AppCompatActivity;
import android.os.Bundle;
import android.util.Log;
import go.Seq;
import com.example.game.yourgamemobile.EbitenView;

public class MainActivity extends AppCompatActivity {

    private static final String TAG = "MainActivity";
    private EbitenView ebitenView;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        // try {
        // Log.d(TAG, "Setting Seq context");
        // Seq.setContext(getApplicationContext());
        // Log.d(TAG, "Creating EbitenView");
        // ebitenView = new EbitenView(this);
        // Log.d(TAG, "Setting content view");
        // setContentView(ebitenView);
        // Log.d(TAG, "onCreate completed successfully");
        // } catch (Exception e) {
        // Log.e(TAG, "Error in onCreate", e);
        // throw e; // Re-throw to crash with full stack trace
        // }
    }

    @Override
    protected void onPause() {
        super.onPause();
        if (ebitenView != null) {
            ebitenView.suspendGame();
        }

    }

    @Override
    protected void onResume() {
        super.onResume();
        if (ebitenView != null) {
            ebitenView.resumeGame();
        }
    }
}