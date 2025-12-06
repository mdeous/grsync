# :camera: grsync

> Effortlessly download photos from Ricoh GR cameras to your computer.

`grsync` is a cross-platform CLI tool that simplifies transferring photos from your Ricoh GR camera to your computer by automating Bluetooth connection and Wi-Fi setup.

**Tested with:** Ricoh GR IIIx (should work with GR II and GR III models as well)

## :information_source: How it works

1. **Discovers cameras** via Bluetooth scanning
2. **Connects** to your camera via Bluetooth
3. **Enables** the camera's integrated Wi-Fi hotspot
4. **Displays** hotspot connection credentials
5. **Waits** for you to connect to the Wi-Fi hotspot
6. **Downloads** photos to your target directory

## :rocket: Installation

### Using go install

```bash
go install github.com/mdeous/grsync@latest
```

### From source

```bash
# Clone the repository
git clone https://github.com/mdeous/grsync.git
cd grsync

# Build the application
go build

# (Optional) Install to your PATH
go install
```

## :clipboard: Usage

`grsync` provides three main commands:

### :mag: Search for Cameras

Scan for available Ricoh GR cameras via Bluetooth:

```bash
./grsync search
```

This will scan for 10 seconds and display all discovered cameras with their names and signal strength.

### :inbox_tray: Sync Photos

Download photos from your camera:

```bash
# Basic usage - downloads all photos to current directory
./grsync sync --camera="GR_XXXXXX"

# Specify a custom download destination
./grsync sync --camera="GR_XXXXXX" --dest="~/Pictures/Ricoh"

# Download only JPG files
./grsync sync --camera="GR_XXXXXX" --extensions="jpg"

# Download only DNG files
./grsync sync --camera="GR_XXXXXX" --extensions="dng"

# Download both JPG and DNG files
./grsync sync --camera="GR_XXXXXX" --extensions="jpg,dng"

# Download all supported file types (default)
./grsync sync --camera="GR_XXXXXX" --extensions="all"

# Short alias
./grsync s --camera="GR_XXXXXX"
```

**Options:**

- `-c, --camera` (required): Name of the camera to connect to
- `-d, --dest`: Destination directory (default: current directory)
- `-e, --extensions`: File types to download - `jpg`, `dng`, or `all` (default: `all`)

### :clipboard: List Photos

List all photos on your camera without downloading:

```bash
./grsync list --camera="GR_XXXXXX"

# Short alias
./grsync l --camera="GR_XXXXXX"
```

**Options:**

- `-c, --camera` (required): Name of the camera to connect to

### :bulb: Tips

- **Finding your camera name**: Use `grsync search` or check your camera's Bluetooth settings. The name usually starts with `GR_` followed by a unique identifier.
- **Existing files**: Photos already present in the destination folder are automatically skipped.
- **Directory structure**: The original folder structure from the camera is preserved during download.

## :handshake: Contributing

Contributions are welcome! Feel free to submit issues or pull requests.

## :copyright: License

This project is open source and available under the [MIT License](LICENSE).
