package embedding

import (
	"context"
	"encoding/binary"
	"hash/fnv"
)

type TextEmbedder interface {
	EmbedText(ctx context.Context, text string) ([]float32, error)
}

type DeterministicTextEmbedder struct{}

func NewDeterministicTextEmbedder() *DeterministicTextEmbedder {
	return &DeterministicTextEmbedder{}
}

func (e *DeterministicTextEmbedder) EmbedText(_ context.Context, text string) ([]float32, error) {
	h := fnv.New64a()
	_, _ = h.Write([]byte(text))
	sum := h.Sum64()

	buffer := make([]byte, 8)
	binary.LittleEndian.PutUint64(buffer, sum)

	out := make([]float32, 8)
	for i := 0; i < 8; i++ {
		value := float32(buffer[i]) / 255.0
		if value == 0 {
			value = 0.001
		}
		out[i] = value
	}

	return out, nil
}

var _ TextEmbedder = (*DeterministicTextEmbedder)(nil)
