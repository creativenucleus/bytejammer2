# Set up directories

```
mkdir _bytejammer-data
mkdir testdir
mkdir tic
```

Put TIC.exe and .dlls in the tic/ directory

# Update config.json

"runnables"
    -> "tic-80-client"
        -> "filepath": "(the full path to)/tic/tic80.exe",

"runnables"
    -> "tic-80-server"
        -> "filepath": "(the full path to)/tic/tic80.exe",

# Run two CLI windows with matching socketurl

bytejammer2.exe kiosk-client --socketurl ws://drone.alkama.com:9000/bytejammer/evoke

bytejammer2.exe kiosk-server --socketurl ws://drone.alkama.com:9000/bytejammer/evoke

# Snapshots

- In the client CLI window, you can press SPACE. 
- That should echo "Sending Snapshot! Snapshot Sent!"
- This will create a new file in testdir/
- The kiosk-server will pick randomly from this directory every 10 seconds