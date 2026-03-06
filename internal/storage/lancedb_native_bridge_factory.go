package storage

import (
	"context"
	"fmt"
)

type nativeLanceDBBridgeFactoryFunc func(ctx context.Context, uri string, tableName string, region string) (lanceDBBridge, error)

var lanceDBNativeBridgeFactory nativeLanceDBBridgeFactoryFunc = func(_ context.Context, _ string, _ string, _ string) (lanceDBBridge, error) {
	return nil, fmt.Errorf("lancedb native backend is disabled in this build; use -tags lancedb_native (with native LanceDB artifacts) or set VIDERA_STORAGE_BACKEND=chromem")
}

func newNativeLanceDBBridge(ctx context.Context, uri string, tableName string, region string) (lanceDBBridge, error) {
	return lanceDBNativeBridgeFactory(ctx, uri, tableName, region)
}
