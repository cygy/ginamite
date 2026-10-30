package image

import "testing"

func TestNewVersions(t *testing.T) {
	tests := []struct {
		name            string
		small, thumb    bool
		wantCount       int
		wantSmallWidth  int
		wantThumbWidth  int
		wantFullName    string
		wantFull2xWidth int
		wantFull2xPath  string
	}{
		{"full only", false, false, 2, 0, 0, "photo.png", 600, "/dest/p/h/photo@2x.png"},
		{"with small", true, false, 4, 100, 0, "photo.png", 600, "/dest/p/h/photo@2x.png"},
		{"with thumb", false, true, 4, 0, 50, "photo.png", 600, "/dest/p/h/photo@2x.png"},
		{"with all", true, true, 6, 100, 50, "photo.png", 600, "/dest/p/h/photo@2x.png"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			versions := NewVersions("/src/Photo.PNG", 300, 150, "/dest", tt.small, tt.thumb)

			if got := len(versions.All()); got != tt.wantCount {
				t.Errorf("len(All()) = %d, want %d", got, tt.wantCount)
			}
			if versions.Full.Size1x.Name != tt.wantFullName {
				t.Errorf("Full.Size1x.Name = %q, want %q", versions.Full.Size1x.Name, tt.wantFullName)
			}
			if versions.Full.Size2x.Width != tt.wantFull2xWidth || versions.Full.Size2x.AbsolutePath != tt.wantFull2xPath {
				t.Errorf("Full.Size2x = %+v, want width %d and path %q", versions.Full.Size2x, tt.wantFull2xWidth, tt.wantFull2xPath)
			}
			if versions.Small.Size1x.Width != tt.wantSmallWidth || versions.Thumb.Size1x.Width != tt.wantThumbWidth {
				t.Errorf("small, thumb width = %d, %d, want %d, %d", versions.Small.Size1x.Width, versions.Thumb.Size1x.Width, tt.wantSmallWidth, tt.wantThumbWidth)
			}
		})
	}
}
