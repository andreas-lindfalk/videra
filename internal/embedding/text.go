package embedding

import (
	"context"
	"hash/fnv"
	"strings"
	"unicode"
)

type TextEmbedder interface {
	EmbedText(ctx context.Context, text string) ([]float32, error)
}

type DeterministicTextEmbedder struct{}

const embeddingDimensions = 8

var semanticCanonicalTokens = map[string]string{
	"cost":         "budget",
	"price":        "budget",
	"pricing":      "budget",
	"spend":        "budget",
	"expense":      "budget",
	"expenses":     "budget",
	"financial":    "budget",
	"finance":      "budget",
	"planning":     "roadmap",
	"plan":         "roadmap",
	"timeline":     "roadmap",
	"milestone":    "roadmap",
	"milestones":   "roadmap",
	"actions":      "actions",
	"steps":        "actions",
	"step":         "actions",
	"summary":      "closing",
	"wrapup":       "closing",
	"wrap":         "closing",
	"introduction": "intro",
	"opening":      "intro",
	"chat":         "discussion",
	"conversation": "discussion",
	"talk":         "discussion",
}

func NewDeterministicTextEmbedder() *DeterministicTextEmbedder {
	return &DeterministicTextEmbedder{}
}

func (e *DeterministicTextEmbedder) EmbedText(_ context.Context, text string) ([]float32, error) {
	tokens := tokenizeAndNormalize(text)
	if len(tokens) == 0 {
		tokens = []string{"empty"}
	}

	out := make([]float32, embeddingDimensions)
	for idx, token := range tokens {
		addHashedFeature(out, token, 1.0)
		if idx > 0 {
			addHashedFeature(out, tokens[idx-1]+"_"+token, 0.6)
		}
	}
	addHashedFeature(out, strings.Join(tokens, " "), 0.25)

	normalizeVector(out)
	return out, nil
}

func tokenizeAndNormalize(text string) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}

	var normalized strings.Builder
	normalized.Grow(len(text))
	for _, r := range strings.ToLower(text) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			normalized.WriteRune(r)
			continue
		}
		normalized.WriteRune(' ')
	}

	parts := strings.Fields(normalized.String())
	if len(parts) == 0 {
		return nil
	}

	for i := range parts {
		parts[i] = normalizeToken(parts[i])
	}

	return parts
}

func normalizeToken(token string) string {
	if canonical, ok := semanticCanonicalTokens[token]; ok {
		return canonical
	}
	return token
}

func addHashedFeature(vector []float32, feature string, weight float32) {
	h := fnv.New64a()
	_, _ = h.Write([]byte(feature))
	sum := h.Sum64()
	bucket := int(sum % uint64(len(vector)))

	sign := float32(1.0)
	if sum&1 == 1 {
		sign = -1.0
	}

	vector[bucket] += sign * weight
}

func normalizeVector(vector []float32) {
	var sumSquares float32
	for _, value := range vector {
		sumSquares += value * value
	}
	if sumSquares == 0 {
		vector[0] = 1
		return
	}

	invNorm := float32(1.0) / sqrt(sumSquares)
	for i := range vector {
		vector[i] *= invNorm
	}
}

func sqrt(value float32) float32 {
	z := value
	for i := 0; i < 6; i++ {
		z -= (z*z - value) / (2 * z)
	}
	return z
}

var _ TextEmbedder = (*DeterministicTextEmbedder)(nil)
