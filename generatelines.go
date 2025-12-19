package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

const (
	defaultWidth = 80
	authorName   = "Christian K. Bjørnsrud"
	repoURL      = "https://github.com/CKB78/GenerateLines"
	version      = "1.0.0"
)

func main() {
	args := os.Args[1:]

	// Version handling
	if len(args) > 0 {
		switch strings.ToLower(strings.TrimSpace(args[0])) {
		case "version", "-v", "--version", "/v":
			fmt.Printf("GenerateLines %s\n", version)
			return
		}
	}

	// Help handling
	if len(args) > 0 {
		switch strings.ToLower(strings.TrimSpace(args[0])) {
		case "/?", "help", "-h", "--help":
			printHelp()
			return
		}
	}

	// Friendly hint when running interactively
	if len(args) == 0 {
		fmt.Println(helpHint())
		fmt.Println()
	}

	lines, filename, overwriteFlag, width, mode, modeArg,
		usedDefaultWidth, usedDefaultMode, err := getArgsOrPrompt(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		fmt.Fprintln(os.Stderr, helpHint())
		os.Exit(1)
	}

	exists := fileExists(filename)
	overwrite := false

	// Use one reader for any interactive prompts in main
	in := bufio.NewReader(os.Stdin)

	if exists {
		if overwriteFlag != "" {
			overwrite = parseYesNo(overwriteFlag)
			if overwrite {
				fmt.Printf("%s already exists. Overwriting...\n", filename)
			} else {
				fmt.Printf("%s already exists. Not overwriting. Exiting.\n", filename)
				return
			}
		} else {
			overwrite, err = promptYesNoR(in, fmt.Sprintf("%s already exists. Overwrite? [y/n]: ", filename))
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
			if !overwrite {
				fmt.Println("Not overwriting. Exiting.")
				return
			}
		}
	}

	openFlag := os.O_CREATE | os.O_WRONLY
	if overwrite {
		openFlag |= os.O_TRUNC
	} else if exists {
		fmt.Println("File exists and overwrite not allowed. Exiting.")
		return
	}

	f, err := os.OpenFile(filename, openFlag, 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error opening file:", err)
		os.Exit(1)
	}
	defer f.Close()

	totalChars := lines * width
	if mode == "pi" {
		fmt.Printf("Mode=pi will generate %d digits (%d lines × %d cols)\n",
			totalChars, lines, width)
	}

	// Build default usage note
	defaultNote := ""
	switch {
	case usedDefaultWidth && usedDefaultMode:
		defaultNote = " [using default width and mode]"
	case usedDefaultWidth:
		defaultNote = " [using default width]"
	case usedDefaultMode:
		defaultNote = " [using default mode]"
	}

	fmt.Printf(
		"Generating %d lines (width=%d, mode=%s)%s -> %s\n",
		lines, width, mode, defaultNote, filename,
	)

	w := bufio.NewWriterSize(f, 1024*64)
	defer w.Flush()

	gen, err := newGenerator(mode, modeArg, totalChars)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	for i := 0; i < lines; i++ {
		line := gen.NextLine(width)
		if _, err := w.WriteString(line + "\n"); err != nil {
			fmt.Fprintln(os.Stderr, "Error writing:", err)
			os.Exit(1)
		}
	}

	fmt.Println("Done!")
}

// helpHint returns the preferred help command hint for the current OS.
func helpHint() string {
	if runtime.GOOS == "windows" {
		return `Tip: run "generatelines /?" for parameters and modes.`
	}
	return `Tip: run "generatelines -h" for parameters and modes.`
}

// printHelp prints command usage, parameters, and available modes to stdout.
func printHelp() {
	fmt.Printf(`GenerateLines v%s
Author: %s
Repository: %s

Generate a text file with N lines of repeatable content.

Usage:
  generatelines <lines> <filename> [y|n] [width] [mode] [modeArg]
  generatelines /?
  generatelines help
  generatelines -h
  generatelines --help
  generatelines version
  generatelines --version

Parameters (positional):
  lines        Number of lines to generate (required unless prompted)
  filename     Output file name (required unless prompted)

Optional parameters:
  y | n        Auto-answer overwrite prompt if file already exists
  width        Line width (columns). Default: 80
  mode         Content generation mode. Default: ascii
  modeArg      Additional argument for selected mode

Modes:
  ascii        Printable ASCII characters (32–126)
  digits       Digits 0–9
  upper        Uppercase letters A–Z
  char         Repeat a single character (requires modeArg)
               Example: generatelines 100 out.txt y 80 char #
  pi           Digits of pi mapped to printable ASCII characters
               Total digits generated = lines × width

Notes:
  - If parameters are omitted, the program will prompt interactively.
  - Defaults are width=80 and mode=ascii.

Examples:
  generatelines 1000 lines.txt
  generatelines 1000 lines.txt y
  generatelines 1000 uppercase.txt y 120 upper
  generatelines 1000 characters.txt y 80 char #
  generatelines 1000 pi.txt n 80 pi
`, version, authorName, repoURL)
}

// getArgsOrPrompt parses positional CLI arguments, or falls back to interactive prompts
// when required arguments are missing. It also reports whether width/mode were chosen
// implicitly (defaults) or explicitly provided by the user.
func getArgsOrPrompt(args []string) (
	lines int,
	filename string,
	overwriteFlag string,
	width int,
	mode string,
	modeArg string,
	usedDefaultWidth bool,
	usedDefaultMode bool,
	err error,
) {

	var linesStr, fileStr string

	width = defaultWidth
	mode = "ascii"
	usedDefaultWidth = true
	usedDefaultMode = true

	// Use one reader for the entire interactive sequence (important for tests and pipes).
	in := bufio.NewReader(os.Stdin)

	if len(args) == 0 {
		linesStr, err = promptLineR(in, "Enter number of lines: ")
		if err != nil {
			return
		}
		fileStr, err = promptLineR(in, "Enter filename: ")
		if err != nil {
			return
		}
	} else if len(args) == 1 {
		linesStr = args[0]
		fileStr, err = promptLineR(in, "Enter filename: ")
		if err != nil {
			return
		}
	} else {
		linesStr = args[0]
		fileStr = args[1]
	}

	lines, err = parsePositiveInt(linesStr)
	if err != nil {
		err = fmt.Errorf(`invalid number of lines: %q (expected a positive integer)`, strings.TrimSpace(linesStr))
		return
	}

	filename = strings.TrimSpace(fileStr)
	if filename == "" {
		err = errors.New("filename cannot be empty")
		return
	}

	rest := []string{}
	if len(args) >= 3 {
		rest = args[2:]
	}

	if len(rest) >= 1 && looksLikeYesNo(rest[0]) {
		overwriteFlag = rest[0]
		rest = rest[1:]
	}

	if len(rest) >= 1 {
		if n, werr := parsePositiveInt(rest[0]); werr == nil {
			width = n
			usedDefaultWidth = false
			rest = rest[1:]
		}
	}

	if len(rest) >= 1 {
		mode = strings.ToLower(strings.TrimSpace(rest[0]))
		usedDefaultMode = false
		rest = rest[1:]
	}

	if len(rest) >= 1 {
		modeArg = rest[0]
	}

	switch mode {
	case "", "ascii":
		mode = "ascii"
	case "digit", "digits":
		mode = "digits"
	case "upper", "uppercase":
		mode = "upper"
	case "char", "character":
		mode = "char"
	case "pi":
		mode = "pi"
	default:
		err = fmt.Errorf("unknown mode: %s", mode)
		return
	}

	if mode == "char" {
		modeArg = strings.TrimSpace(modeArg)
		if modeArg == "" {
			err = errors.New("mode=char requires modeArg")
			return
		}
	}

	if width <= 0 {
		width = defaultWidth
		usedDefaultWidth = true
	}

	return
}

// looksLikeYesNo reports whether s is a valid yes/no token (y/yes/n/no), case-insensitive.
func looksLikeYesNo(s string) bool {
	s = strings.TrimSpace(s)
	return strings.EqualFold(s, "y") || strings.EqualFold(s, "yes") ||
		strings.EqualFold(s, "n") || strings.EqualFold(s, "no")
}

// parsePositiveInt parses s as a positive integer (> 0).
func parsePositiveInt(s string) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0, err
	}
	if n <= 0 {
		return 0, errors.New("must be > 0")
	}
	return n, nil
}

// promptLineR prints a prompt and reads a single line from r.
func promptLineR(r *bufio.Reader, prompt string) (string, error) {
	fmt.Print(prompt)

	text, err := r.ReadString('\n')
	if err != nil {
		// Accept EOF if we got some input (e.g. piped input without trailing newline)
		if len(text) == 0 {
			return "", err
		}
	}

	return strings.TrimSpace(text), nil
}

// promptYesNoR prompts the user until a yes/no answer is provided (y/yes/n/no), case-insensitive.
func promptYesNoR(r *bufio.Reader, prompt string) (bool, error) {
	for {
		s, err := promptLineR(r, prompt)
		if err != nil {
			return false, err
		}
		if strings.EqualFold(s, "y") || strings.EqualFold(s, "yes") {
			return true, nil
		}
		if strings.EqualFold(s, "n") || strings.EqualFold(s, "no") {
			return false, nil
		}
		fmt.Println("Please answer y or n.")
	}
}

// parseYesNo reports whether s is a "yes" token (y/yes), case-insensitive.
func parseYesNo(s string) bool {
	return strings.EqualFold(strings.TrimSpace(s), "y") ||
		strings.EqualFold(strings.TrimSpace(s), "yes")
}

// fileExists reports whether the given file path exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Generator produces fixed-width lines of content for output files.
type Generator interface {
	NextLine(width int) string
}

// newGenerator constructs a Generator for the given mode.
// totalChars is used for sizing when mode requires precomputation (e.g. pi).
func newGenerator(mode, modeArg string, totalChars int) (Generator, error) {
	switch mode {
	case "ascii":
		return &cycleGen{palette: []byte(buildAsciiSequence())}, nil

	case "digits":
		return &cycleGen{palette: []byte("0123456789")}, nil

	case "upper":
		return &cycleGen{palette: []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")}, nil

	case "char":
		modeArg = strings.TrimSpace(modeArg)
		if modeArg == "" {
			return nil, errors.New("mode=char requires modeArg")
		}
		r := []rune(modeArg)
		if len(r) == 0 {
			return nil, errors.New("mode=char requires modeArg")
		}
		return &singleCharGen{ch: string(r[0])}, nil

	case "pi":
		if totalChars <= 0 {
			totalChars = 1
		}
		palette := []byte(buildAsciiSequence())
		return &piGen{
			palette: palette,
			spigot:  newPiSpigot(totalChars),
		}, nil

	default:
		return nil, errors.New("unknown mode")
	}
}

// cycleGen emits characters by cycling through a fixed palette.
type cycleGen struct {
	palette []byte
	pos     int
}

// NextLine returns the next line of output with the given width.
func (g *cycleGen) NextLine(width int) string {
	out := make([]byte, width)
	for i := 0; i < width; i++ {
		out[i] = g.palette[g.pos%len(g.palette)]
		g.pos++
	}
	return string(out)
}

// singleCharGen emits a line consisting of a single repeated character.
type singleCharGen struct {
	ch string
}

func (g *singleCharGen) NextLine(width int) string {
	return strings.Repeat(g.ch, width)
}

// piGen emits digits of π mapped onto a printable ASCII palette.
type piGen struct {
	palette []byte
	spigot  *piSpigot
}

func (g *piGen) NextLine(width int) string {
	out := make([]byte, width)
	for i := 0; i < width; i++ {
		out[i] = g.palette[g.spigot.NextDigit()%len(g.palette)]
	}
	return string(out)
}

// buildAsciiSequence returns printable ASCII characters (32..126) as a string.
func buildAsciiSequence() string {
	var b strings.Builder
	for i := 32; i <= 126; i++ {
		b.WriteByte(byte(i))
	}
	return b.String()
}

// piSpigot implements a base-10 spigot algorithm for streaming digits of π.
type piSpigot struct {
	a        []int
	queue    []int
	nines    int
	predigit int
	started  bool
}

// newPiSpigot creates a spigot sized to generate at least the given number of digits.
func newPiSpigot(digits int) *piSpigot {
	size := digits*10/3 + 1
	a := make([]int, size)
	for i := range a {
		a[i] = 2
	}
	return &piSpigot{a: a, queue: make([]int, 0, 32)}
}

// NextDigit returns the next digit of π (0..9).
func (p *piSpigot) NextDigit() int {
	if len(p.queue) > 0 {
		d := p.queue[0]
		p.queue = p.queue[1:]
		return d
	}

	for {
		q := 0
		for i := len(p.a) - 1; i >= 0; i-- {
			x := 10*p.a[i] + q*(i+1)
			den := 2*(i+1) - 1
			p.a[i] = x % den
			q = x / den
		}
		p.a[0] = q % 10
		q /= 10

		switch q {
		case 9:
			p.nines++
		case 10:
			p.queue = append(p.queue, p.predigit+1)
			for i := 0; i < p.nines; i++ {
				p.queue = append(p.queue, 0)
			}
			p.predigit = 0
			p.nines = 0
		default:
			p.queue = append(p.queue, p.predigit)
			for i := 0; i < p.nines; i++ {
				p.queue = append(p.queue, 9)
			}
			p.predigit = q
			p.nines = 0
		}

		for len(p.queue) > 0 {
			if !p.started && p.queue[0] == 0 {
				p.queue = p.queue[1:]
				continue
			}
			p.started = true
			d := p.queue[0]
			p.queue = p.queue[1:]
			return d
		}
	}
}
