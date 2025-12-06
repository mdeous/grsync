# :camera: grsync

> Effortlessly download photos from Ricoh GR cameras to your computer.

`grsync` is a cross-platform tool simplifies the process of transferring photos from your Ricoh GR camera to your computer by automating the connection to the camera and the download process

Tested with a Ricoh GR IIIx only, but should work with GR II and GR III models as well.

## :information_source: How it works

1. Connects to the camera via Bluetooth
2. Enables the camera integrated Wi-Fi hotspot
3. Displays the hotspot connection info
4. Waits for the user to connect to the hotspot
5. Enumerates photos stored on the camera
6. Downloads photos to the target directory

## :rocket: Installation

```bash
# Clone the repository
git clone https://github.com/mdeous/grsync.git
cd grsync

# Build the application
go build
```

## :clipboard: Usage

```bash
# Basic usage with required camera name
./grsync --camera="GR_XXXXXX"

# Specify a custom download destination (defaults to current directory)
./grsync --camera="GR_XXXXXX" --dest="~/Pictures/Ricoh"

# Download only JPG files
./grsync --camera="GR_XXXXXX" --ext="jpg"

# Download only DNG files
./grsync --camera="GR_XXXXXX" --ext="dng"

# Download both JPG and DNG files
./grsync --camera="GR_XXXXXX" --ext="jpg,dng"

# Download all supported photo types (default behavior)
./grsync --camera="GR_XXXXXX" --ext="all"

# Show help
./grsync --help
```

> :bulb: **Tip**: The camera name usually starts with "GR\_" followed by a unique identifier. Check your camera's Bluetooth settings to find the exact name.

## :construction: TODO

- Find a way to automate Wi-Fi hotspot connection
- Parallelized photos download

## :handshake: Contributing

Contributions are welcome! Feel free to submit issues or pull requests.

## :scroll: License

This project is open source and available under the [MIT License](LICENSE).
