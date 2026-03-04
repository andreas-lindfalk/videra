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
