
# Purpose
This tool downloads the files from the server concurrently

# Usage
```bash
go build 
cd ~/desiredLocation/
~/binary/location/download <ip address of the server>:8000
``` 

It connects to the server at /status/ and creates the required directory structure required for the files.
Then for each file it crates (max) 50 connections and downloads parts of data.
