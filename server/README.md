# Building on Raspberry Pi
The following go env variables must be set to the following to correctly build and run the project:
CGO_ENABLED=1
GOOS=linux
GOARCH=arm64
CC=aarch64-linux-gnu-gcc (or another arm64 compatible compiler)

# Resources
Used the following to fix a bluetooth hid permission issue on linux:
https://unix.stackexchange.com/questions/85379/dev-hidraw-read-permissions
Used the following documentation for extracting Joycon data:
https://github.com/dekuNukem/Nintendo_Switch_Reverse_Engineering/blob/master/bluetooth_hid_notes.md