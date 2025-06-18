# Nitro

**Nitro** is a fast, multi-threaded download accelerator written in Go. It boosts download speeds by splitting files into parts and downloading them in parallel. Nitro supports both **HTTP** and **FTP** protocols.

## Features

- Accelerated downloads using concurrent connections
- Supports **HTTP**, and **FTP**
- Smart default parallelism based on your CPU
- Single self-contained binary
- Simple command-line interface

## Installation

The recommended way to install Nitro is to download a tarball from the [Releases](https://github.com/terslang/nitro/releases) page.

### 1. Download, verify and build

1. Download the appropriate binary for your platform.
2. (Optional) Verify the GPG signature:
    You need to have my public key in your keyring to verify the signature. Import my key by running...
    
    ```bash
    gpg --recv-key C8B2A95D8D855A9D8C6F0C78BCBCAE31ECE05007
    ```
    
    Verify
    
    ```bash
    gpg --verify nitro-<version>.tar.gz.sig nitro-<version>.tar.gz
    ```
    
    

3. Extract and move the binary to a directory in your `PATH`, for example:

    ```bash
    tar -xzf nitro-<version>.tar.gz
    cd nitro-<version>
    go build -o nitro ./cmd/nitro
    sudo mv nitro /usr/local/bin/
    ```

### 2. Build from source (not recommended for most users)

You can build Nitro from source if needed:

```bash
git clone https://github.com/terslang/nitro.git
cd nitro
go build -o nitro ./cmd/nitro
```

> You can add an optional `-ldflags "-s"` option to the `go build` command to drop the debug symbols resulting in a smaller binary.

> Note: Nitro has dependencies, but the final binary is self-contained and requires no shared libraries or runtime dependencies.

## Usage

```bash
nitro [options] URL
```

### Examples
```bash
nitro https://example.com/movie.mp4
```
```bash
nitro -p 8 -o movie.mp4 https://example.com/movie.mp4
```
```bash
nitro -p 8 -o movie.mp4 ftp://username:password@example.com/movie.mp4
```

This downloads `movie.mp4` using 8 parallel connections.

### CLI Options

| Option             | Description                                                   |
|--------------------|---------------------------------------------------------------|
| `--parallel`, `-p` | Number of concurrent connections to use (default: auto-tuned) |
| `--output`, `-o`   | Output file name (default: `"output-file"`)                   |
| `--verbose`        | Enable debug logs                                             |
| `--help`, `-h`     | Show help                                                     |
| `--version`, `-v`  | Show version                                                  |

## Protocol Support

- ✅ HTTP
- ✅ FTP

## License

GPL-3.0-only

---

> Nitro is under active development. Feedback and contributions are welcome!
