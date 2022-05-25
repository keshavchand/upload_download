# Purpose
This tool starts the server that the downloader will use to download the data from

# Usage
```bash
go build 
cd ~/desiredLocation/
~/binary/location/upload
``` 

This starts a server on port 8000 and exposes two endpoints /status/ and /download/.
/status/ shows all the files and folders in the current directory and /download/ serves them.
