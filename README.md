# [M5Burner](https://github.com/matu6968/m5burner) update server

This is a replacement update server meant to replace the original update server found at https://m5burner-cdn.m5stack.com

It offers other architecture support then just x64 on the official version by M5Stack, emulates the officlal update server endpoints and that's all

How to use:

1. Install [Go](https://go.dev/doc/install) if not already installed
2. Clone this repo
4. Run go build -o m5burner-update-server
5. (optional) To set a custom web server port: Make  .env file with the contents: PORT=<any_available_port>
6. Make the update folder structure like here:
```
patches/
  windows/
  darwin/
  linux/
```
7. Put your patch files like shown here:
```
patches/
  windows/
    202403201500-windows-x64.zip
    202403201500-windows-x86.zip
    ...
  darwin/
    202403201500-darwin-x64.zip
    202403201500-darwin-arm64.zip
    ...
  linux/
    202403201500-linux-x64.zip
    202403201500-linux-arm64.zip
    ...
```    
8. Run the binary with ./m5burner-update-server

API documentation:

```
<server_url>/appVersion.info - get latest version info (for x64)

query parameters:
?platform=<linux,windows,darwin> - set platform type (required)
?arch=<arm64,armv7l,x86,x64> - get latest version info (for selected architecture, optional)

If unable to read latest version info from file: HTTP status 500, response: Error reading patches

If version info does not exist: HTTP status 404, response: No updates found

If version info has invalid architecture specified in query parameter: HTTP status 404, response: Invalid architecture

If version info has no platform specified in query parameter: HTTP status 400, response: Platform parameter is required

If version info has invalid platform specified in query parameter: HTTP status 404, response: Invalid platform

If version info does not have a update file found for the target platform/architecture specified in query parameter: HTTP status 404, response: Update file not found

<server_url>/patch/<yyyymmddhhmm>-<platform>.zip - get latest patch for specified platform (for x64)

query parameters:
?arch=<arm64,armv7l,x86,x64> - get latest patch for specified platform (and for selected architecture, optional)
?timestamp=<unix_timestamp> - unused part (other then checking the timestamp) by the core part of the server but gets defined by the M5Burner client (optional)

If patch path is invalid: HTTP status 400, response: Invalid path

If patch file name path does not match regex syntax: HTTP status 400, response: Invalid file format

If patch path has invalid platform specified in the file name: HTTP status 400, response: Invalid platform

If version info has invalid timestamp specified in query parameter: HTTP status 400, response: Invalid timestamp

If unable to read latest file: HTTP status 500, response: Error reading patches

If patch path does not exist for the desired platform/architecture: HTTP status 404, response: Update file not found
```
