# Bouncing Ball Physics Game

A simple physics-based bouncing ball game built with [Ebiten](https://ebiten.org/), featuring gravity, damping, and drag-to-throw mechanics. Supports multiple platforms: Desktop, Web (WebAssembly), and Android.

## Features

- Realistic physics simulation with gravity and damping
- Drag-to-throw ball interaction
- Multi-platform support (Desktop, Web, Android)
- Touch input support for mobile devices

## Prerequisites

### System Requirements
- Go 1.24.7 or later
- Git

### For Desktop Development
- No additional requirements beyond Go

### For Web Development (WebAssembly)
- No additional requirements beyond Go

### For Android Development
- Android SDK Platform 33
- Android SDK Command Line Tools (latest)
- Android NDK 25.2.9519653
- Java 8 JDK (for Gradle compatibility)

## Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/brensch/game.git
   cd game
   ```

2. **Install Go dependencies:**
   ```bash
   go mod download
   ```

### Android Setup (Optional)

If you want to build for Android, you'll need to set up the Android development environment:

1. **Install Android SDK:**
   - Download and install Android Studio, or
   - Use command line tools only

2. **Set up environment variables:**
   ```bash
   export ANDROID_HOME=~/Android/Sdk
   export ANDROID_NDK_HOME=$ANDROID_HOME/ndk/25.2.9519653
   export PATH=$PATH:$ANDROID_HOME/cmdline-tools/latest/bin:$ANDROID_HOME/platform-tools:$ANDROID_NDK_HOME
   ```

3. **Install required SDK components:**
   ```bash
   # Accept licenses
   yes | sdkmanager --licenses

   # Install platform and build tools
   sdkmanager "platform-tools" "platforms;android-33" "build-tools;33.0.2"
   ```

4. **Install gomobile (for Android binding):**
   ```bash
   go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest
   ```

## Building and Running

### Desktop

```bash
go run main.go
```

### Web (WebAssembly)

```bash
# Build for web
go run github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest bind -target js -o game.wasm github/brensch/game/cmd/mobile

# Serve locally (requires a web server)
python3 -m http.server 8080
# Then open http://localhost:8080 in your browser
```

### Android

1. **Generate Android AAR:**
   ```bash
   ./build_aar.sh
   ```

2. **Build APK:**
   ```bash
   cd android_app
   ./gradlew assembleDebug
   ```

3. **Install on device:**
   ```bash
   adb install app/build/outputs/apk/debug/app-debug.apk
   ```

## Project Structure

```
game/
├── main.go                 # Desktop entry point
├── go.mod                  # Go module definition
├── cmd/
│   └── mobile/            # Mobile-specific code
├── pkg/
│   └── game/              # Game logic package
│       └── game.go        # Main game implementation
├── android_app/           # Android project
│   ├── app/
│   │   ├── build.gradle   # Android app configuration
│   │   ├── libs/          # AAR dependencies
│   │   └── src/main/java/com/example/game/
│   │       └── MainActivity.java  # Android activity
│   └── gradle/wrapper/    # Gradle wrapper
├── index.html             # Web deployment template
└── README.md              # This file
```

## Game Controls

- **Desktop:** Click and drag the ball to throw it
- **Mobile:** Touch and drag the ball to throw it

The ball will bounce with realistic physics including gravity and velocity damping.

## Development

### Modifying Game Physics

Edit `pkg/game/game.go` to adjust:
- `Gravity`: Downward acceleration
- `Damping`: Velocity reduction on bounces
- `ThrowFactor`: Sensitivity of throw force
- `BallRadius`: Size of the ball

### Adding Platforms

The game uses Ebiten's cross-platform capabilities. To add new platforms:

1. Use `ebitenmobile bind` with appropriate `-target` flag
2. Follow platform-specific integration guides in the [Ebiten documentation](https://ebiten.org/documents/mobile.html)

### Hot Reload (Development)

For faster development, you can use [Air](https://github.com/cosmtrek/air) for hot reloading:

```bash
go install github.com/cosmtrek/air@latest
air
```

## Troubleshooting

### Android Build Issues

- **SDK licenses not accepted:** Run `yes | sdkmanager --licenses`
- **Missing NDK:** Ensure `ANDROID_NDK_HOME` points to a valid NDK installation
- **Gradle errors:** Check Java version compatibility (requires Java 8 for this project)

### WebAssembly Issues

- **CORS errors:** Serve files from a local web server, not directly from file://
- **Browser compatibility:** Ensure your browser supports WebAssembly

### General Issues

- **Go version:** Ensure you're using Go 1.24.7 or later
- **Dependencies:** Run `go mod tidy` if you encounter import issues

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test on multiple platforms
5. Submit a pull request

## License

This project is open source. See LICENSE file for details.

## Credits

Built with [Ebiten](https://ebiten.org/) - A dead simple 2D game engine for Go.