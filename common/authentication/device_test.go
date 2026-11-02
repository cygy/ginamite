package authentication

import "testing"

func TestDevice_SetUpNameAndDetails(t *testing.T) {
	const chrome = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	const bot = "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"

	tests := []struct {
		name        string
		device      Device
		wantName    string
		wantDetails string
	}{
		{"web browser", Device{Name: chrome, Type: TypeWeb}, "Chrome", "Intel Mac OS X 10_15_7"},
		{"web bot", Device{Name: bot, Type: TypeWeb}, "Googlebot", "bot"},
		{"ios keeps its name", Device{Name: "iPhone de Cyril", Type: TypeIOS}, "iPhone de Cyril", ""},
		{"android keeps its name", Device{Name: "Pixel", Type: TypeAndroid, Details: "x"}, "Pixel", "x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := tt.device.Name
			tt.device.SetUpNameAndDetails()

			if tt.device.RawName != raw {
				t.Errorf("RawName = %q, want %q", tt.device.RawName, raw)
			}
			if tt.device.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", tt.device.Name, tt.wantName)
			}
			if tt.device.Details != tt.wantDetails {
				t.Errorf("Details = %q, want %q", tt.device.Details, tt.wantDetails)
			}
		})
	}
}
