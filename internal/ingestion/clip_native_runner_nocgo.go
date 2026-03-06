//go:build !cgo

package ingestion

import (
	"context"
	"fmt"
)

type NativeCLIPRunner struct{}

func NewNativeCLIPRunner() *NativeCLIPRunner {
	return &NativeCLIPRunner{}
}

func (r *NativeCLIPRunner) Prepare(_ string, _ string) error {
	return fmt.Errorf("native clip backend requires a CGO-enabled build")
}

func (r *NativeCLIPRunner) EmbedFrame(_ context.Context, _ string, _ string, _ int) ([]float32, error) {
	return nil, fmt.Errorf("native clip backend requires a CGO-enabled build")
}
