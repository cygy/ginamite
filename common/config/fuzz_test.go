package config

import "testing"

// FuzzUnmarshal checks that no configuration file, even malformed, makes the loading panic.
func FuzzUnmarshal(f *testing.F) {
	f.Add([]byte("server:\n  port: 8080\njwt:\n  secret: !env JWT_SECRET\n"))
	f.Add([]byte("cookie:\n  same_site: none\n"))
	f.Add([]byte("a: [\n"))

	f.Fuzz(func(t *testing.T, source []byte) {
		c := NewConfiguration()
		if err := unmarshal(source, c, &c.sections); err == nil {
			_ = c.Cookie.validate()
		}
	})
}
