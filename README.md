# GenerateLines

GenerateLines is a CLI utility written in Go that generates text files with a specified number of lines and a fixed line width. It’s useful for testing editors, import pipelines, log processors, and tools that need predictable, repeatable text input.

## Features

- Cross-platform (Windows, macOS, Linux)
- Interactive mode when required parameters are omitted
- Fixed line width (default: 80 columns)
- Multiple output modes: `ascii`, `digits`, `upper`, `char`, `pi`
- Safe overwrite handling (prompted unless explicitly provided)

## Installation

### Prerequisites

- Go installed and available on your `PATH` (Go 1.20+ recommended)
- If you use the optional “install system-wide” step on Linux/macOS:
  - You need permission to write to the target directory (typically via `sudo`)
  - The target directory (e.g. `/usr/local/bin`) should exist and be on your `PATH`

### Build from source

#### Linux / macOS

Build a local binary in the current directory:

```bash
go build -o generatelines .
```

Optionally install system-wide:

```bash
sudo install -m 0755 generatelines /usr/local/bin/generatelines
```

> Note: On some macOS setups (especially Homebrew on Apple Silicon), you may prefer `/opt/homebrew/bin` instead of `/usr/local/bin` if that’s what’s on your `PATH`.

#### Alternative: install to your Go bin directory

If you prefer not to use `sudo`, you can install to your Go bin directory:

```bash
go install .
```

The binary will be placed in your Go bin directory (commonly `$(go env GOPATH)/bin`). Make sure that directory is on your `PATH`.

#### Windows

```bat
go build -o generatelines.exe .
```

## Usage

```text
generatelines <lines> <filename> [y|n] [width] [mode] [modeArg]
```

Help:

```text
generatelines /?
generatelines -h
generatelines --help
```

Version:

```text
generatelines --version
generatelines version
```

## Modes

- `ascii`  
  Printable ASCII characters (32–126)

- `digits`  
  Digits `0–9`

- `upper`  
  Uppercase letters `A–Z`

- `char`  
  Repeat a single character (requires `modeArg`)

- `pi`  
  Digits of π (π) as **raw numeric characters (`0–9`)**  
  Total digits generated = `lines × width`

  Optional modeArg:
  - `ascii`  
    Map π digits (`0–9`) to printable ASCII characters (legacy behavior)

## Examples

Generate 1000 lines using defaults (80 columns, `ascii`):

```bash
generatelines 1000 lines.txt
```

Overwrite if the file already exists:

```bash
generatelines 1000 lines.txt y
```

Uppercase with custom width:

```bash
generatelines 500 out.txt y 120 upper
```

Single repeated character:

```bash
generatelines 200 hashes.txt y 80 char #
```

π digits (default pi behavior):

```bash
generatelines 100 pi_digits.txt y 80 pi
```

π digits mapped to printable ASCII characters:

```bash
generatelines 100 pi_ascii.txt y 80 pi ascii
```

## Fun fact

This utility was originally written to answer a very practical question:  
**how many lines of text the new Microsoft Edit could handle** before things started getting interesting.

## Author

Christian K. Bjørnsrud  
Repository: https://github.com/CKB78/GenerateLines

## License

MIT
