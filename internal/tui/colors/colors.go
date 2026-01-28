package colors

import "github.com/charmbracelet/lipgloss"

// Adaptive color definitions for light/dark terminal themes
// Using CompleteColor with ANSI fallback for better compatibility
var (
	// Hint - Gray for hints and dim text (ANSI 8 = bright black/gray)
	Hint = lipgloss.CompleteColor{
		TrueColor: "#9B9B9B",
		ANSI256:   "244", // gray
		ANSI:      "8",   // bright black
	}
	
	// Success - Green for success messages (ANSI 2 = green)
	Success = lipgloss.CompleteColor{
		TrueColor: "#00A000",
		ANSI256:   "34", // dark green
		ANSI:      "2",  // green
	}
	
	// Error - Red for error messages (ANSI 1 = red)
	Error = lipgloss.CompleteColor{
		TrueColor: "#FF0000",
		ANSI256:   "196", // red
		ANSI:      "1",   // red
	}
	
	// Info - Blue for info messages (ANSI 4 = blue)
	Info = lipgloss.CompleteColor{
		TrueColor: "#0000CD",
		ANSI256:   "20", // dark blue
		ANSI:      "4",  // blue
	}
	
	// Warning - Yellow for warnings (ANSI 3 = yellow)
	Warning = lipgloss.CompleteColor{
		TrueColor: "#B8860B",
		ANSI256:   "136", // dark goldenrod
		ANSI:      "3",   // yellow
	}
	
	// HeaderBG - Purple for header background (ANSI 5 = magenta)
	HeaderBG = lipgloss.CompleteColor{
		TrueColor: "#7D56F4",
		ANSI256:   "99", // purple
		ANSI:      "5",  // magenta
	}
	
	// HeaderFG - White for header text (ANSI 7 = white)
	HeaderFG = lipgloss.CompleteColor{
		TrueColor: "#FFFFFF",
		ANSI256:   "231", // white
		ANSI:      "7",   // white
	}
	
	// Selected - Purple for selected items (ANSI 5 = magenta)
	Selected = lipgloss.CompleteColor{
		TrueColor: "#7D56F4",
		ANSI256:   "99", // purple
		ANSI:      "5",  // magenta
	}
	
	// Section - Magenta for section headers (ANSI 5 = magenta)
	Section = lipgloss.CompleteColor{
		TrueColor: "#9B30FF",
		ANSI256:   "93", // purple
		ANSI:      "5",  // magenta
	}
)
