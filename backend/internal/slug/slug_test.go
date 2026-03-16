package slug

import "testing"

func TestGenerate(t *testing.T) {
	tests := []struct {
		title    string
		fallback string
		want     string
	}{
		{"Sunset at the Beach", "abc12345", "sunset-at-the-beach"},
		{"", "abc12345", "abc12345"},
		{"  Spaces  Everywhere  ", "abc12345", "spaces-everywhere"},
		{"UPPERCASE TITLE", "abc12345", "uppercase-title"},
		{"Special! @#$ Characters", "abc12345", "special-characters"},
		{"---leading-trailing---", "abc12345", "leading-trailing"},
		{"Portra 400 — Night Shots", "abc12345", "portra-400-night-shots"},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			got := Generate(tt.title, tt.fallback)
			if got != tt.want {
				t.Errorf("Generate(%q, %q) = %q, want %q", tt.title, tt.fallback, got, tt.want)
			}
		})
	}
}
