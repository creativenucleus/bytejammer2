# ByteJammer v2

This is a rebuild / experimental build. I'll almost certainly eventually collapse it into the [ByteJammer v1](https://github.com/creativenucleus/bytejammer2) repo.

This project is **NOT** ready for outside contributions just now, but that time will come!  

[Latest Code and documentation - GitHub](https://github.com/creativenucleus/bytejammer2)

For celebration of the TIC-80 livecoding / effects scene.

- A jukebox / robot VJ that plays TIC-80 effects for personal enjoyment.
- A standalone client-server for running ByteJams.

Please read the documentation before running.

### **IMPORTANT**

This is a work in progress **USE AT YOUR OWN RISK**. It's currently early days, so things are more likely to go wrong, but [so far] nobody has reported any ill effects.

Features should currently be considered experimental, and liable to change, possibly in a way that is not backward compatible. The format of arguments for the CLI is likely to be particularly in flux, and this documentation may lag behind development (I tend to code chunks, and periodically review docs). Please feel welcome to contact me if you have questions.

## Setup

A config.json file is required.  

An example _config.json file is provided. Please remove the underscore from the filename and edit that file as appropriate.

Please note: Some functionality requires a [custom version of the Bytebattle build](https://github.com/creativenucleus/TIC-80-bytebattle) that permits reading AND writing from the filesystem (the standard version permits reading OR writing).

### Example config.json

```json
{
    "work_dir": "./_bytejammer-data",
    "control_panel": {
        "port": 9000
    },
    "runnables": {
        "tic-80-client": {
            "filepath": "./tics/old-bytejammer-build/tic80-win.exe",
            "args": [
                "--skip"
            ]
        },
        "tic-80-server": {
            "filepath": "./tics/old-bytejammer-build/tic80-win.exe",
            "args": [
                "--skip",
                "--scale=2"
            ]
        }
    },
    "jukebox": {
        "rotate_period_in_seconds": 15
    }
}
```

## Running a ByteWall

Put `attractmode.lua` in the `_bytejammer-data/kiosk-server-playlist` folder (you might have to run once for that folder to be created).

```cli
.\bytejammer2.exe kiosk-server --connection host --port 8900 --endpoint /kiosk/listener --obs-overlay-port 4000

.\bytejammer2.exe kiosk-client --url ws://localhost:8900/kiosk/listener --startercodepath ./startercode.lua
```
(this launches a webpanel available at http://localhost:9000, or at the port specified in the config.json)

## Goals

- Decent functionality from low configuration  
Ideally, the UI to do things at a basic level should be as light / easy to start as possible.
- Compatibility  
There are a number of tools in the livecoding/TIC/websocket ecosystem. Aim to be compatible.
- Augmentation  
Ensure the ByteJammer adds value to the ecosystem.
- Composability  
Providing a selection of components that can be glued together to allow unanticipated interactions.

## Ideas for Components to Implement

- Abstraction of connections between things so that they may be transparently websocket or internal.
- MessagePipe  
Flows one way
- ExeLauncher  
Decide whether to bundle executables (e.g. websocket-compliant versions of TIC), as well as having them linkable.
- FileObserver  
Reads a file every X seconds
Broadcast a Message on Change  
- FileProvider  
An abstraction of a file system?