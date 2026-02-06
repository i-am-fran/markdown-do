package colors

import "github.com/charmbracelet/lipgloss"

// Modern color palette inspired by Mole's clean design
// Using CompleteColor with ANSI fallback for better compatibility
var (
	// Hint - Subtle gray for hints and dim text
	Hint = lipgloss.CompleteColor{
		TrueColor: "#6B7280", // Modern gray
		ANSI256:   "242",     // gray
		ANSI:      "8",       // bright black
	}

	// Success - Modern green for success messages
	Success = lipgloss.CompleteColor{
		TrueColor: "#10B981", // Emerald green
		ANSI256:   "42",       // green
		ANSI:      "2",        // green
	}

	// Error - Modern red for error messages
	Error = lipgloss.CompleteColor{
		TrueColor: "#EF4444", // Modern red
		ANSI256:   "196",     // red
		ANSI:      "1",       // red
	}

	// Info - Modern blue for info messages
	Info = lipgloss.CompleteColor{
		TrueColor: "#3B82F6", // Modern blue
		ANSI256:   "33",      // blue
		ANSI:      "4",       // blue
	}

	// Warning - Modern yellow for warnings
	Warning = lipgloss.CompleteColor{
		TrueColor: "#F59E0B", // Amber
		ANSI256:   "214",     // orange-yellow
		ANSI:      "3",       // yellow
	}

	// HeaderBG - Modern accent color for header background
	HeaderBG = lipgloss.CompleteColor{
		TrueColor: "#6366F1", // Indigo
		ANSI256:   "63",      // indigo
		ANSI:      "5",       // magenta
	}

	// HeaderFG - White for header text
	HeaderFG = lipgloss.CompleteColor{
		TrueColor: "#FFFFFF",
		ANSI256:   "231", // white
		ANSI:      "7",   // white
	}

	// Selected - Accent color for selected items
	Selected = lipgloss.CompleteColor{
		TrueColor: "#6366F1", // Indigo (matches header)
		ANSI256:   "63",      // indigo
		ANSI:      "5",       // magenta
	}

	// Section - Softer purple for section headers
	Section = lipgloss.CompleteColor{
		TrueColor: "#8B5CF6", // Purple
		ANSI256:   "93",      // purple
		ANSI:      "5",       // magenta
	}

	// Border - Subtle border color
	Border = lipgloss.CompleteColor{
		TrueColor: "#E5E7EB", // Light gray
		ANSI256:   "250",     // light gray
		ANSI:      "7",       // white
	}

	// Accent - Primary accent color for highlights
	Accent = lipgloss.CompleteColor{
		TrueColor: "#6366F1", // Indigo
		ANSI256:   "63",      // indigo
		ANSI:      "5",       // magenta
	}

	// Muted - Muted text color
	Muted = lipgloss.CompleteColor{
		TrueColor: "#9CA3AF", // Muted gray
		ANSI256:   "247",     // light gray
		ANSI:      "8",       // bright black
	}

	// Background - Subtle background color for panels
	Background = lipgloss.CompleteColor{
		TrueColor: "#F9FAFB", // Very light gray
		ANSI256:   "255",     // white
		ANSI:      "7",       // white
	}
)
