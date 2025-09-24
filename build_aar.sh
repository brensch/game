#!/bin/bash

# Build script for Ebiten Android AAR

set -e

echo "Building Ebiten AAR for Android..."

# Set Android environment variables if not set
if [ -z "$ANDROID_HOME" ]; then
    export ANDROID_HOME=~/Android/Sdk
fi

if [ -z "$ANDROID_NDK_HOME" ]; then
    export ANDROID_NDK_HOME=$ANDROID_HOME/ndk/25.2.9519653
fi

export PATH=$PATH:$ANDROID_HOME/cmdline-tools/latest/bin:$ANDROID_HOME/platform-tools:$ANDROID_NDK_HOME

# Check if ebitenmobile is installed
if ! command -v ebitenmobile &> /dev/null; then
    echo "Installing ebitenmobile..."
    go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest
fi

# Build the AAR
echo "Building AAR..."
ebitenmobile bind -target android -javapkg com.example.game -o game.aar github/brensch/game/cmd/mobile

echo "AAR built successfully: game.aar"

# Copy to Android app
echo "Copying AAR to Android app..."
cp game.aar android_app/app/libs/

echo "Done. You can now build the Android app with ./gradlew assembleDebug"