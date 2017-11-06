package gettext

// Locale is like a dictionary.
type Locale interface {
	// Load .mo file and parse.
	Load() error
	// Get translation from dictionary.
	Get(string) string
}
