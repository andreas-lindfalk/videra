//go:build lancedb_native

package storage

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/apache/arrow/go/v17/arrow"
	"github.com/apache/arrow/go/v17/arrow/array"
	"github.com/apache/arrow/go/v17/arrow/memory"
	"github.com/lancedb/lancedb-go/pkg/contracts"
	"github.com/lancedb/lancedb-go/pkg/lancedb"
)

const lanceDBCloudURIPrefix = "db://"

type lanceDBNativeBridge struct {
	conn      contracts.IConnection
	tableName string
	mu        sync.Mutex
}

func init() {
	lanceDBNativeBridgeFactory = newNativeLanceDBBridgeImpl
}

func newNativeLanceDBBridgeImpl(ctx context.Context, uri string, tableName string, region string) (lanceDBBridge, error) {
	var connOptions *contracts.ConnectionOptions
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(uri)), lanceDBCloudURIPrefix) {
		trimmedRegion := strings.TrimSpace(region)
		if trimmedRegion == "" {
			return nil, fmt.Errorf("VIDERA_LANCEDB_REGION is required when VIDERA_LANCEDB_URI uses db://")
		}
		connOptions = &contracts.ConnectionOptions{Region: &trimmedRegion}
	}

	conn, err := lancedb.Connect(ctx, uri, connOptions)
	if err != nil {
		return nil, fmt.Errorf("connect lancedb: %w", err)
	}

	return &lanceDBNativeBridge{conn: conn, tableName: tableName}, nil
}

func (b *lanceDBNativeBridge) UpsertSegments(ctx context.Context, rows []lanceDBSegmentRow) error {
	if len(rows) == 0 {
		return nil
	}

	embeddingDim := len(rows[0].Embedding)
	if embeddingDim == 0 {
		return fmt.Errorf("missing embedding dimension")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	table, err := b.openOrCreateTable(ctx, embeddingDim)
	if err != nil {
		return err
	}
	defer func() { _ = table.Close() }()

	record, err := rowsToArrowRecord(rows, embeddingDim)
	if err != nil {
		return fmt.Errorf("build arrow record: %w", err)
	}
	defer record.Release()

	if err := table.AddRecords(ctx, []arrow.Record{record}, nil); err != nil {
		return fmt.Errorf("add records to lancedb table: %w", err)
	}

	return nil
}

func (b *lanceDBNativeBridge) SearchSegments(ctx context.Context, queryEmbedding []float32, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 5
	}
	if len(queryEmbedding) == 0 {
		return []map[string]any{}, nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	table, err := b.conn.OpenTable(ctx, b.tableName)
	if err != nil {
		if isMissingTableError(err) {
			return []map[string]any{}, nil
		}
		return nil, fmt.Errorf("open lancedb table: %w", err)
	}
	defer func() { _ = table.Close() }()

	rows, err := table.VectorSearch(ctx, lanceDBFieldEmbedding, queryEmbedding, limit)
	if err != nil {
		return nil, fmt.Errorf("vector search lancedb table: %w", err)
	}

	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		mapped := make(map[string]any, len(row))
		for key, value := range row {
			mapped[key] = value
		}
		out = append(out, mapped)
	}

	return out, nil
}

func (b *lanceDBNativeBridge) Reset(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.conn.DropTable(ctx, b.tableName); err != nil && !isMissingTableError(err) {
		return fmt.Errorf("drop lancedb table: %w", err)
	}
	return nil
}

func (b *lanceDBNativeBridge) openOrCreateTable(ctx context.Context, embeddingDim int) (contracts.ITable, error) {
	table, err := b.conn.OpenTable(ctx, b.tableName)
	if err == nil {
		return table, nil
	}
	if !isMissingTableError(err) {
		return nil, fmt.Errorf("open lancedb table: %w", err)
	}

	schema, err := lancedb.NewSchema(newLanceDBArrowSchema(embeddingDim))
	if err != nil {
		return nil, fmt.Errorf("create lancedb schema: %w", err)
	}

	table, err = b.conn.CreateTable(ctx, b.tableName, schema)
	if err != nil {
		return nil, fmt.Errorf("create lancedb table: %w", err)
	}
	return table, nil
}

func newLanceDBArrowSchema(embeddingDim int) *arrow.Schema {
	fields := []arrow.Field{
		{Name: lanceDBFieldDocID, Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: lanceDBFieldVideoID, Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: lanceDBFieldFilePath, Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: lanceDBFieldStartMs, Type: arrow.PrimitiveTypes.Int64, Nullable: false},
		{Name: lanceDBFieldEndMs, Type: arrow.PrimitiveTypes.Int64, Nullable: false},
		{Name: lanceDBFieldType, Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: lanceDBFieldSourcePath, Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: lanceDBFieldText, Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: lanceDBFieldEmbedding, Type: arrow.FixedSizeListOf(int32(embeddingDim), arrow.PrimitiveTypes.Float32), Nullable: false},
	}
	return arrow.NewSchema(fields, nil)
}

func rowsToArrowRecord(rows []lanceDBSegmentRow, embeddingDim int) (arrow.Record, error) {
	schema := newLanceDBArrowSchema(embeddingDim)
	recordBuilder := array.NewRecordBuilder(memory.NewGoAllocator(), schema)
	defer recordBuilder.Release()

	docIDBuilder := recordBuilder.Field(0).(*array.StringBuilder)
	videoIDBuilder := recordBuilder.Field(1).(*array.StringBuilder)
	filePathBuilder := recordBuilder.Field(2).(*array.StringBuilder)
	startMsBuilder := recordBuilder.Field(3).(*array.Int64Builder)
	endMsBuilder := recordBuilder.Field(4).(*array.Int64Builder)
	typeBuilder := recordBuilder.Field(5).(*array.StringBuilder)
	sourcePathBuilder := recordBuilder.Field(6).(*array.StringBuilder)
	textBuilder := recordBuilder.Field(7).(*array.StringBuilder)
	embeddingBuilder := recordBuilder.Field(8).(*array.FixedSizeListBuilder)
	embeddingValues := embeddingBuilder.ValueBuilder().(*array.Float32Builder)

	for _, row := range rows {
		docIDBuilder.Append(row.DocID)
		videoIDBuilder.Append(row.VideoID)
		filePathBuilder.Append(row.FilePath)
		startMsBuilder.Append(row.StartMs)
		endMsBuilder.Append(row.EndMs)
		typeBuilder.Append(row.Type)
		sourcePathBuilder.Append(row.SourcePath)
		textBuilder.Append(row.Text)

		embeddingBuilder.Append(true)
		for idx := 0; idx < embeddingDim; idx++ {
			if idx < len(row.Embedding) {
				embeddingValues.Append(row.Embedding[idx])
				continue
			}
			embeddingValues.Append(0)
		}
	}

	return recordBuilder.NewRecord(), nil
}

func isMissingTableError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(message, "failed to open table") ||
		strings.Contains(message, "not found") ||
		strings.Contains(message, "no such") ||
		strings.Contains(message, "does not exist")
}
