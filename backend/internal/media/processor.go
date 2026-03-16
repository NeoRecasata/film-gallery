package media

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"

	"github.com/buckket/go-blurhash"
	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
)

type Variant struct {
	Data   []byte
	Width  int
	Height int
}

type Processor struct {
	thumbWidth  int
	mediumWidth int
	fullWidth   int
}

func NewProcessor(thumbWidth, mediumWidth, fullWidth int) *Processor {
	return &Processor{thumbWidth: thumbWidth, mediumWidth: mediumWidth, fullWidth: fullWidth}
}

func (p *Processor) GenerateVariants(r io.Reader) (map[string]*Variant, error) {
	src, err := imaging.Decode(r, imaging.AutoOrientation(true))
	if err != nil {
		return nil, fmt.Errorf("decoding image: %w", err)
	}
	sizes := map[string]int{"thumb": p.thumbWidth, "medium": p.mediumWidth, "full": p.fullWidth}
	variants := make(map[string]*Variant, 3)
	for name, maxWidth := range sizes {
		resized := src
		if src.Bounds().Dx() > maxWidth {
			resized = imaging.Resize(src, maxWidth, 0, imaging.Lanczos)
		}
		var buf bytes.Buffer
		if err := webp.Encode(&buf, resized, &webp.Options{Quality: 85}); err != nil {
			return nil, fmt.Errorf("encoding %s to webp: %w", name, err)
		}
		bounds := resized.Bounds()
		variants[name] = &Variant{Data: buf.Bytes(), Width: bounds.Dx(), Height: bounds.Dy()}
	}
	return variants, nil
}

func (p *Processor) ExtractDimensions(r io.Reader) (int, int, error) {
	cfg, _, err := image.DecodeConfig(r)
	if err != nil {
		return 0, 0, fmt.Errorf("decoding image config: %w", err)
	}
	return cfg.Width, cfg.Height, nil
}

func (p *Processor) GenerateBlurHash(r io.Reader) (string, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return "", fmt.Errorf("decoding image for blurhash: %w", err)
	}
	hash, err := blurhash.Encode(4, 3, img)
	if err != nil {
		return "", fmt.Errorf("generating blurhash: %w", err)
	}
	return hash, nil
}
