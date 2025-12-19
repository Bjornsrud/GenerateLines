package main

import (
	"os"
	"strings"
	"testing"
)

func TestBuildAsciiSequence(t *testing.T) {
	s := buildAsciiSequence()
	if len(s) != (126-32+1) {
		t.Fatalf("expected ascii palette length 95, got %d", len(s))
	}
	if s[0] != byte(32) || s[len(s)-1] != byte(126) {
		t.Fatalf("expected palette to start at 32 and end at 126, got %d..%d", s[0], s[len(s)-1])
	}
}

func TestNewGenerator_UnknownMode(t *testing.T) {
	_, err := newGenerator("bananas", "", 100)
	if err == nil {
		t.Fatalf("expected error for unknown mode, got nil")
	}
}

func TestNewGenerator_CharModeRequiresArg(t *testing.T) {
	// Note: in CLI parsing, mode=char without arg is rejected earlier.
	// This test focuses on generator behavior with empty arg.
	_, err := newGenerator("char", "", 100)
	if err == nil {
		t.Fatalf("expected error for char mode without arg, got nil")
	}
}

func TestGenerator_ASCII(t *testing.T) {
	g, err := newGenerator("ascii", "", 1000)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	width := 80
	palette := []byte(buildAsciiSequence())
	if len(palette) == 0 {
		t.Fatal("ascii palette is empty")
	}

	line1 := g.NextLine(width)
	if len(line1) != width {
		t.Fatalf("expected length %d, got %d", width, len(line1))
	}

	// Ensure all chars are printable ASCII (32..126)
	for i := 0; i < len(line1); i++ {
		if line1[i] < 32 || line1[i] > 126 {
			t.Fatalf("non-printable ascii at %d: %d", i, line1[i])
		}
	}

	// Verify deterministic cycling:
	// line1 should start at palette[0] and line2 should start at palette[width].
	if line1[0] != palette[0] {
		t.Fatalf("expected first char %q, got %q", palette[0], line1[0])
	}

	line2 := g.NextLine(width)
	if len(line2) != width {
		t.Fatalf("expected length %d, got %d", width, len(line2))
	}

	want2First := palette[width%len(palette)]
	if line2[0] != want2First {
		t.Fatalf("expected second line first char %q, got %q", want2First, line2[0])
	}

	// Optional: also check that a known index matches expected palette advancement.
	// line1[i] should equal palette[i]
	checkIdx := 10
	if line1[checkIdx] != palette[checkIdx%len(palette)] {
		t.Fatalf("expected line1[%d]=%q, got %q", checkIdx, palette[checkIdx%len(palette)], line1[checkIdx])
	}
}


func TestGenerator_Digits(t *testing.T) {
	g, err := newGenerator("digits", "", 1000)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	line := g.NextLine(50)
	if len(line) != 50 {
		t.Fatalf("expected length 50, got %d", len(line))
	}
	for i := 0; i < len(line); i++ {
		if line[i] < '0' || line[i] > '9' {
			t.Fatalf("expected digit at %d, got %q", i, line[i])
		}
	}
}

func TestGenerator_Upper(t *testing.T) {
	g, err := newGenerator("upper", "", 1000)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	line := g.NextLine(52)
	if len(line) != 52 {
		t.Fatalf("expected length 52, got %d", len(line))
	}
	for i := 0; i < len(line); i++ {
		if line[i] < 'A' || line[i] > 'Z' {
			t.Fatalf("expected A-Z at %d, got %q", i, line[i])
		}
	}
}

func TestGenerator_Char(t *testing.T) {
	g, err := newGenerator("char", "#", 1000)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	line := g.NextLine(33)
	if line != strings.Repeat("#", 33) {
		t.Fatalf("unexpected char line: %q", line)
	}
}

func TestPiSpigot_FirstDigits(t *testing.T) {
	// Known first digits of pi: 3 1 4 1 5 9 2 6 5 3 5 8 9 7 9 3
	want := []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5, 8, 9, 7, 9, 3}

	s := newPiSpigot(len(want))
	for i, wd := range want {
		got := s.NextDigit()
		if got != wd {
			t.Fatalf("pi digit %d: expected %d, got %d", i, wd, got)
		}
	}
}

func TestGenerator_PiMode_MappingToAsciiPalette(t *testing.T) {
	// In pi mode, digits 0..9 are mapped to printable ASCII palette by index.
	// Palette index 0 corresponds to ASCII 32 (space), 1 -> '!', etc.
	// So digit '3' maps to palette[3] = ASCII 35 '#'
	g, err := newGenerator("pi", "", 80) // totalChars used only for spigot sizing
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	line := g.NextLine(10)
	if len(line) != 10 {
		t.Fatalf("expected length 10, got %d", len(line))
	}

	// First digits of pi: 3,1,4,1,5,9,2,6,5,3
	// Expected chars: palette[d] where palette[0] = ' ' (32)
	palette := []byte(buildAsciiSequence())
	wantDigits := []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3}
	want := make([]byte, 10)
	for i, d := range wantDigits {
		want[i] = palette[d%len(palette)]
	}

	if line != string(want) {
		t.Fatalf("unexpected pi-mapped line.\nwant: %q\ngot:  %q", string(want), line)
	}
}

func TestGetArgsOrPrompt_DefaultFlags_WhenOmitted(t *testing.T) {
	// Only required args -> defaults should be used (width + mode)
	lines, filename, ow, width, mode, modeArg, defW, defM, err := getArgsOrPrompt([]string{"10", "out.txt"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if lines != 10 || filename != "out.txt" || ow != "" || width != defaultWidth || mode != "ascii" || modeArg != "" {
		t.Fatalf("unexpected parsed result: lines=%d file=%q ow=%q width=%d mode=%q arg=%q",
			lines, filename, ow, width, mode, modeArg)
	}
	if !defW || !defM {
		t.Fatalf("expected defW=true and defM=true, got defW=%v defM=%v", defW, defM)
	}
}

func TestGetArgsOrPrompt_NoDefaultFlags_WhenUserSpecifiesDefaults(t *testing.T) {
	// User explicitly sets width=80 and mode=ascii -> should NOT be marked as default usage
	lines, filename, ow, width, mode, modeArg, defW, defM, err := getArgsOrPrompt([]string{"10", "out.txt", "80", "ascii"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if lines != 10 || filename != "out.txt" || ow != "" || width != 80 || mode != "ascii" || modeArg != "" {
		t.Fatalf("unexpected parsed result: lines=%d file=%q ow=%q width=%d mode=%q arg=%q",
			lines, filename, ow, width, mode, modeArg)
	}
	if defW || defM {
		t.Fatalf("expected defW=false and defM=false, got defW=%v defM=%v", defW, defM)
	}
}

func TestGetArgsOrPrompt_OverwriteFlag_CaseInsensitive(t *testing.T) {
	lines, filename, ow, width, mode, _, defW, defM, err := getArgsOrPrompt([]string{"10", "out.txt", "Y"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if lines != 10 || filename != "out.txt" || strings.ToLower(ow) != "y" {
		t.Fatalf("unexpected overwrite flag parsing: lines=%d file=%q ow=%q", lines, filename, ow)
	}
	if width != defaultWidth || mode != "ascii" || !defW || !defM {
		t.Fatalf("unexpected defaults: width=%d mode=%q defW=%v defM=%v", width, mode, defW, defM)
	}
}

func TestGetArgsOrPrompt_ModeChar_RequiresModeArg(t *testing.T) {
	_, _, _, _, _, _, _, _, err := getArgsOrPrompt([]string{"10", "out.txt", "80", "char"})
	if err == nil {
		t.Fatalf("expected error for char mode without modeArg")
	}
}

func TestGetArgsOrPrompt_InteractivePrompts(t *testing.T) {
	// Simulate interactive input: lines=7, filename=test.txt
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	tmp, err := os.CreateTemp("", "stdin-*")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	_, _ = tmp.WriteString("7\n")
	_, _ = tmp.WriteString("test.txt\n")
	_, _ = tmp.Seek(0, 0)

	os.Stdin = tmp

	lines, filename, ow, width, mode, modeArg, defW, defM, err := getArgsOrPrompt([]string{})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if lines != 7 || filename != "test.txt" || ow != "" || width != defaultWidth || mode != "ascii" || modeArg != "" {
		t.Fatalf("unexpected interactive result: lines=%d file=%q ow=%q width=%d mode=%q arg=%q",
			lines, filename, ow, width, mode, modeArg)
	}
	if !defW || !defM {
		t.Fatalf("expected defaults for width+mode in interactive flow")
	}
}
