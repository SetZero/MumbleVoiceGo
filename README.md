# Mumble Voice Go

Mumble Voice Go is a Mumble Voice Server / Gateway which can be accessed via WebSocket and/or REST API to play sound on a Mumble Server

# Installation

1. You need to have libopus and libopusfile installed on your system

    You can do this under **Ubuntu/Debian/...** with:
    ```bash
    sudo apt-get install pkg-config libopus-dev libopusfile-dev
    ```
    **Mac:**

    ```bash
    brew install pkg-config opus opusfile
    ```

    **Windows:**
    
    Visit [Opus downloads](https://opus-codec.org/downloads/) and download libopus. Put the DLL under a Linkable Directory
    
2. Next you can just do:

    ```bash
    go build
    ```

