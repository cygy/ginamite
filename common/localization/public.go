package localization

// Initialize : initalizes the translation engine.
func Initialize(path string, locales []string) {
	Main = NewEngine()
	Main.Initialize(path, locales)
}

// Translate : returns the right translate function.
func Translate(locale string) TranslateFunc {
	return Main.Translate(locale)
}

// Source : returns the raw source string.
func Source(locale, key string) string {
	return Main.Source(locale, key)
}
