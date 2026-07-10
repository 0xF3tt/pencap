package main

import "os"

const (
	clrBlue  = "\x1b[38;2;122;162;247m" // #7aa2f7
	clrPurp  = "\x1b[38;2;187;154;247m" // #bb9af7
	clrCyan  = "\x1b[38;2;125;207;255m" // #7dcfff
	clrFg    = "\x1b[38;2;192;202;245m" // #c0caf5
	clrDim   = "\x1b[38;2;84;92;126m"   // #545c7e
	clrBold  = "\x1b[1m"
	clrReset = "\x1b[0m"
)

func colorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	info, err := os.Stderr.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func banner() string {
	if !colorEnabled() {
		return "pencap\npentest evidence · findings · report\n\n"
	}
	return clrBlue + "▍" + clrPurp + "▍" + clrCyan + "▍" + clrReset +
		" " + clrBold + clrFg + "pencap" + clrReset + "\n" +
		clrDim + "  pentest evidence · findings · report" + clrReset + "\n\n"
}
