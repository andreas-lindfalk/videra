package storage

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/andreas-lindfalk/videra/internal/embedding"
)

const lanceDBCompatibilityDirName = "lancedb-compat"

type LanceDBStoreOptions struct {
	SplitSharedStorage bool
}

type LanceDBStore struct {
	*ChromemStore
}

var _ VectorStore = (*LanceDBStore)(nil)

func NewLanceDBStore(dataDir string, textEmbedder embedding.TextEmbedder) (*LanceDBStore, error) {
	return NewLanceDBStoreWithOptions(dataDir, textEmbedder, LanceDBStoreOptions{})
}

func NewLanceDBStoreWithOptions(dataDir string, textEmbedder embedding.TextEmbedder, options LanceDBStoreOptions) (*LanceDBStore, error) {
	normalizedDataDir := strings.TrimSpace(dataDir)
	if normalizedDataDir == "" {
		normalizedDataDir = "./data"
	}

	compatibilityDataDir := filepath.Join(normalizedDataDir, lanceDBCompatibilityDirName)
	store, err := NewChromemStoreWithOptions(
		compatibilityDataDir,
		textEmbedder,
		ChromemStoreOptions{SplitSharedStorage: options.SplitSharedStorage},
	)
	if err != nil {
		return nil, fmt.Errorf("initialize lancedb compatibility store: %w", err)
	}

	return &LanceDBStore{ChromemStore: store}, nil
}
