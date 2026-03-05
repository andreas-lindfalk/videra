package embedding

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeterministicTextEmbedderStable(t *testing.T) {
	ctx := context.Background()
	embedder := NewDeterministicTextEmbedder()

	a, err := embedder.EmbedText(ctx, "budget roadmap")
	require.NoError(t, err)
	b, err := embedder.EmbedText(ctx, "budget roadmap")
	require.NoError(t, err)

	require.Equal(t, a, b)
	require.Len(t, a, 8)
}

func TestDeterministicTextEmbedderDistinctInputs(t *testing.T) {
	ctx := context.Background()
	embedder := NewDeterministicTextEmbedder()

	a, err := embedder.EmbedText(ctx, "budget roadmap")
	require.NoError(t, err)
	b, err := embedder.EmbedText(ctx, "incident timeline")
	require.NoError(t, err)

	require.NotEqual(t, a, b)
}

func TestDeterministicTextEmbedderSemanticNeighboring(t *testing.T) {
	ctx := context.Background()
	embedder := NewDeterministicTextEmbedder()

	anchor, err := embedder.EmbedText(ctx, "budget roadmap next actions")
	require.NoError(t, err)

	nearby, err := embedder.EmbedText(ctx, "cost planning next steps")
	require.NoError(t, err)

	far, err := embedder.EmbedText(ctx, "sunset beach vacation")
	require.NoError(t, err)

	nearSimilarity := cosine(anchor, nearby)
	farSimilarity := cosine(anchor, far)

	require.Greater(t, nearSimilarity, farSimilarity)
}

func cosine(left, right []float32) float64 {
	if len(left) == 0 || len(right) == 0 || len(left) != len(right) {
		return 0
	}

	var dot float64
	for i := range left {
		dot += float64(left[i] * right[i])
	}

	return dot
}
