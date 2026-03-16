package media

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
)

func createTestJPEG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80})
	return buf.Bytes()
}

func TestProcessor_GenerateVariants(t *testing.T) {
	p := NewProcessor(400, 1200, 2400)
	src := createTestJPEG(3000, 2000)
	variants, err := p.GenerateVariants(bytes.NewReader(src))
	if err != nil {
		t.Fatalf("GenerateVariants: %v", err)
	}
	if len(variants) != 3 {
		t.Fatalf("expected 3 variants, got %d", len(variants))
	}
	if variants["thumb"].Width != 400 {
		t.Errorf("thumb width = %d, want 400", variants["thumb"].Width)
	}
	if variants["medium"].Width != 1200 {
		t.Errorf("medium width = %d, want 1200", variants["medium"].Width)
	}
	if variants["full"].Width != 2400 {
		t.Errorf("full width = %d, want 2400", variants["full"].Width)
	}
}

func TestProcessor_GenerateVariants_SmallImage(t *testing.T) {
	p := NewProcessor(400, 1200, 2400)
	src := createTestJPEG(300, 200)
	variants, err := p.GenerateVariants(bytes.NewReader(src))
	if err != nil {
		t.Fatalf("GenerateVariants: %v", err)
	}
	for name, v := range variants {
		if v.Width != 300 {
			t.Errorf("variant %q width = %d, want 300 (no upscale)", name, v.Width)
		}
	}
}

func TestProcessor_ExtractDimensions(t *testing.T) {
	p := NewProcessor(400, 1200, 2400)
	src := createTestJPEG(1920, 1080)
	w, h, err := p.ExtractDimensions(bytes.NewReader(src))
	if err != nil {
		t.Fatalf("ExtractDimensions: %v", err)
	}
	if w != 1920 || h != 1080 {
		t.Errorf("got %dx%d, want 1920x1080", w, h)
	}
}

func TestProcessor_GenerateBlurHash(t *testing.T) {
	p := NewProcessor(400, 1200, 2400)
	src := createTestJPEG(400, 300)
	hash, err := p.GenerateBlurHash(bytes.NewReader(src))
	if err != nil {
		t.Fatalf("GenerateBlurHash: %v", err)
	}
	if len(hash) == 0 {
		t.Error("blur hash is empty")
	}
}
