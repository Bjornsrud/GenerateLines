# GenerateLines

GenerateLines is a CLI utility written in Go that generates text files with a specified number of lines and a fixed line width, and with various optional types of generated content. 

<img width="873" height="245" alt="image" src="https://github.com/user-attachments/assets/5030dff6-0889-43e5-bfd4-3b6f8e3475de" />

## Features

- Interactive mode when required parameters are omitted
- Fixed line width (default: 80 columns)
- Multiple output modes: `ascii`, `digits`, `upper`, `char`, `pi`
- Safe overwrite handling (prompted unless explicitly provided)

## Installation

### Requirements

Go 1.20+

### Build from source

#### Linux / macOS

Build a local binary in the current directory:

```bash
go build -o generatelines .
```

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
  Digits of PI (π) as **raw numeric characters (`0–9`)**  
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

This utility was initially written to answer a very practical question:  
**how many lines of text can the new Microsoft Edit handle** before things started getting interesting.
Then things got a bit out of hand. :D

## Author

Christian K. Bjørnsrud  
Repository: https://github.com/CKB78/GenerateLines

## License

MIT
