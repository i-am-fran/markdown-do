package markdown

import (
	"strings"
	"sync"

	"github.com/charmbracelet/glamour"
)

var (
	renderer     *glamour.TermRenderer
	rendererOnce sync.Once
	cache        = make(map[string]string)
	cacheMutex   sync.RWMutex
)

// initRenderer initializes the glamour renderer
func initRenderer() {
	rendererOnce.Do(func() {
		var err error
		renderer, err = glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(80),
		)
		if err != nil {
			// Fallback to no styling if renderer fails
			renderer = nil
		}
	})
}

// Render renders markdown text to styled output
func Render(text string) string {
	initRenderer()

	// Check cache first
	cacheMutex.RLock()
	if cached, ok := cache[text]; ok {
		cacheMutex.RUnlock()
		return cached
	}
	cacheMutex.RUnlock()

	// If renderer is nil, return plain text
	if renderer == nil {
		return text
	}

	// Render the markdown
	out, err := renderer.Render(text)
	if err != nil {
		return text
	}

	// Clean up output
	rendered := strings.TrimSpace(out)

	// Cache the result
	cacheMutex.Lock()
	// Limit cache size to prevent memory growth
	if len(cache) > 100 {
		// Clear cache when it gets too large
		cache = make(map[string]string)
	}
	cache[text] = rendered
	cacheMutex.Unlock()

	return rendered
}

// ClearCache clears the markdown rendering cache
func ClearCache() {
	cacheMutex.Lock()
	cache = make(map[string]string)
	cacheMutex.Unlock()
}
