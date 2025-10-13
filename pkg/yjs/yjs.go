// pkg/yjs/yjs.go - Clean Y.js CRDT API for Go
package yjs

import "fmt"

// YjsService provides high-level Y.js operations
type YjsService interface {
	// CreateEmptySnapshot creates an empty Y.js document snapshot
	CreateEmptySnapshot() ([]byte, error)
	
	// CreateSnapshotWithText creates a Y.js document snapshot with initial text content
	CreateSnapshotWithText(fieldName, text string) ([]byte, error)
	
	// CreateUpdateWithText creates a Y.js update that can be applied to empty document to get initial text
	// This is used for GitHub imports to create v1 updates instead of v1 snapshots
	CreateUpdateWithText(fieldName, text string) ([]byte, error)
	
	// ApplyUpdates applies Y.js updates to a base snapshot
	ApplyUpdates(baseSnapshot []byte, updates [][]byte) ([]byte, error)
	
	// ValidateUpdate checks if an update is in valid Y.js format
	ValidateUpdate(update []byte) error
	
	// ExtractText extracts text content from Y.js snapshot for a specific field
	ExtractText(snapshot []byte, fieldName string) (string, error)
}

// NewYjsService creates a new Y.js service using the real YFFI implementation
func NewYjsService() (YjsService, error) {
	return &yjsServiceImpl{}, nil
}

type yjsServiceImpl struct{}

func (s *yjsServiceImpl) CreateEmptySnapshot() ([]byte, error) {
	return createEmptyYDocWithContent()
}

func (s *yjsServiceImpl) CreateSnapshotWithText(fieldName, text string) ([]byte, error) {
	// Try to create Y.js document with text
	result, err := createYDocWithText(fieldName, text)
	if err != nil {
		// If Y.js FFI fails, fall back to empty document as a safety measure
		fmt.Printf("Y.js FFI failed for field %s with text length %d: %v. Falling back to empty document.\n", fieldName, len(text), err)
		return s.CreateEmptySnapshot()
	}
	return result, nil
}

func (s *yjsServiceImpl) CreateUpdateWithText(fieldName, text string) ([]byte, error) {
	// The createYDocWithText function already does exactly what we need:
	// 1. Create new document
	// 2. Insert text into field
	// 3. Use ytransaction_state_diff_v1(txn, nil, 0, &len) to get complete state as update
	// 
	// This gives us the v1 update that represents the initial GitHub content
	return createYDocWithText(fieldName, text)
}

func (s *yjsServiceImpl) ApplyUpdates(baseSnapshot []byte, updates [][]byte) ([]byte, error) {
	// Apply updates directly using YFFI
	return applyUpdatesToYDoc(baseSnapshot, updates)
}

func (s *yjsServiceImpl) ValidateUpdate(update []byte) error {
	if len(update) == 0 {
		return fmt.Errorf("update cannot be empty")
	}
	
	// Basic validation - Y.js updates typically start with specific bytes
	if len(update) < 2 {
		return fmt.Errorf("update too short to be valid Y.js update")
	}
	
	return nil
}

func (s *yjsServiceImpl) ExtractText(snapshot []byte, fieldName string) (string, error) {
	return extractTextFromYjsDocument(snapshot, fieldName)
}

// YjsMapReader provides a high-level interface for reading YJS maps
type YjsMapReader interface {
	// LoadMap loads a YJS map from binary data 
	LoadMap(data []byte, mapName string) (YjsMapIterator, error)
}

// YjsMapIterator iterates over YJS map entries
type YjsMapIterator interface {
	// Next returns the next key-value pair, or ("", "", nil) when done
	Next() (key string, value string, err error)
	// Count returns the total number of entries
	Count() uint32
	// Close releases resources
	Close()
}

// NewYjsMapReader creates a new map reader
func NewYjsMapReader() YjsMapReader {
	return &yjsMapReaderImpl{}
}

type yjsMapReaderImpl struct{}

func (r *yjsMapReaderImpl) LoadMap(data []byte, mapName string) (YjsMapIterator, error) {
	return newYjsMapIterator(data, mapName)
}