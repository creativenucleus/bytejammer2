# ByteJammer (rebuild / experimental)

## Example config.json

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

```cli
.\bytejammer2.exe kiosk-server --connection host --port 8900 --endpoint /kiosk/listener

.\bytejammer2.exe kiosk-client --url ws://localhost:8900/kiosk/listener

(optional --startercodepath a/path/to/a/lua/file.lua)
```
(this launches a webpanel available at http://localhost:9000)

## Goals

- Composability
- Compatibility
- Augmentation

## Components

- MessagePipe
    - Flows one way
- ExeLauncher
- FileObserver
    - Reads a file every X seconds
    - Broadcast Message on Change
- FileProvider