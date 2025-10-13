package yjs

/*
#cgo CFLAGS: -I../../lib
#cgo LDFLAGS: ${SRCDIR}/../../lib/libyrs.a -lpthread -ldl -lm
#include <stdlib.h>
#include <stdint.h>

// Forward declarations to match libyrs.h API
typedef struct YDoc{} YDoc;
typedef void* YTransaction;
typedef void* YText;
typedef void* Branch;
typedef struct YInput YInput;
typedef struct YMapIter{} YMapIter;
typedef struct YOutput YOutput;
typedef struct YMapEntry {
  const char *key;
  const struct YOutput *value;
} YMapEntry;

// Function declarations from libyrs.h - corrected to match actual signatures
extern YDoc* ydoc_new(void);
extern void ydoc_destroy(YDoc* doc);
extern YTransaction* ydoc_write_transaction(YDoc* doc, uint32_t origin_len, const char *origin);
extern YTransaction* ydoc_read_transaction(YDoc* doc);
extern void ytransaction_commit(YTransaction* txn);
extern Branch* ytext(YDoc* doc, const char *name);
extern unsigned char* ytransaction_state_vector_v1(YTransaction* txn, int *len);
extern unsigned char* ytransaction_state_diff_v1(YTransaction* txn, const unsigned char *sv, int sv_len, int *len);
extern void ybinary_destroy(unsigned char *ptr, int len);
extern int ytransaction_apply(YTransaction* txn, const unsigned char *update, int len);
extern void ytext_insert(const Branch* txt, YTransaction* txn, uint32_t index, const char *value, const struct YInput *attrs);
extern char* ytext_string(const Branch* txt, const YTransaction* txn);
extern void ystring_destroy(char* str);
extern Branch* ymap(YDoc* doc, const char *name);
extern YMapIter* ymap_iter(const Branch* map, const YTransaction* txn);
extern void ymap_iter_destroy(YMapIter* iter);
extern struct YMapEntry* ymap_iter_next(YMapIter* iter);
extern void ymap_entry_destroy(struct YMapEntry* entry);
extern uint32_t ymap_len(const Branch* map, const YTransaction* txn);
extern Branch* youtput_read_ymap(const struct YOutput* val);
extern Branch* youtput_read_ytext(const struct YOutput* val);
extern char* youtput_read_string(const struct YOutput* val);
extern const char* youtput_read_binary(const struct YOutput* val);
extern void youtput_destroy(struct YOutput* val);
extern struct YOutput* ymap_get(const Branch* map, const YTransaction* txn, const char* key);
extern YDoc* youtput_read_ydoc(const struct YOutput* val);
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// createEmptyYDocWithContent creates a Y.js document with an empty 'content' text field
func createEmptyYDocWithContent() ([]byte, error) {
	// Create a new Y.js document
	doc := C.ydoc_new()
	if doc == nil {
		return nil, fmt.Errorf("failed to create Y.js document")
	}
	defer C.ydoc_destroy(doc)

	// Get the 'content' text field
	contentName := C.CString("content")
	defer C.free(unsafe.Pointer(contentName))

	textField := C.ytext(doc, contentName)
	if textField == nil {
		return nil, fmt.Errorf("failed to create content text field")
	}

	// Start a write transaction (origin_len=0, origin=NULL for no origin)
	txn := C.ydoc_write_transaction(doc, 0, nil)
	if txn == nil {
		return nil, fmt.Errorf("failed to create write transaction")
	}
	defer C.ytransaction_commit(txn)

	// Get the document content as a Y.js update (this includes the empty 'content' field)
	var updateLen C.int
	updateBytes := C.ytransaction_state_diff_v1(txn, nil, 0, &updateLen)
	if updateBytes == nil {
		return nil, fmt.Errorf("failed to get document content")
	}
	defer C.ybinary_destroy(updateBytes, updateLen)

	// Convert C bytes to Go slice
	goBytes := C.GoBytes(unsafe.Pointer(updateBytes), updateLen)
	result := make([]byte, len(goBytes))
	copy(result, goBytes)

	return result, nil
}

// createYDocWithText creates a Y.js document with initial text content in the specified field
func createYDocWithText(fieldName, text string) ([]byte, error) {
	if fieldName == "" {
		return nil, fmt.Errorf("fieldName cannot be empty")
	}

	// Create a new Y.js document
	doc := C.ydoc_new()
	if doc == nil {
		return nil, fmt.Errorf("failed to create Y.js document")
	}
	defer C.ydoc_destroy(doc)

	// Get the text field (creates if doesn't exist)
	cFieldName := C.CString(fieldName)
	if cFieldName == nil {
		return nil, fmt.Errorf("failed to create C string for field name")
	}
	defer C.free(unsafe.Pointer(cFieldName))

	textField := C.ytext(doc, cFieldName)
	if textField == nil {
		return nil, fmt.Errorf("failed to create text field '%s'", fieldName)
	}

	// Start a write transaction
	txn := C.ydoc_write_transaction(doc, 0, nil)
	if txn == nil {
		return nil, fmt.Errorf("failed to create write transaction")
	}

	// Insert text into the text field if provided
	if text != "" {
		cText := C.CString(text)
		if cText == nil {
			C.ytransaction_commit(txn)
			return nil, fmt.Errorf("failed to create C string for text content")
		}
		defer C.free(unsafe.Pointer(cText))

		C.ytext_insert(textField, txn, 0, cText, nil)
	}

	// Get the document content as Y.js update before committing
	var updateLen C.int
	updateBytes := C.ytransaction_state_diff_v1(txn, nil, 0, &updateLen)

	// Commit the transaction
	C.ytransaction_commit(txn)

	if updateBytes == nil {
		return nil, fmt.Errorf("failed to get document content")
	}
	defer C.ybinary_destroy(updateBytes, updateLen)

	// Convert C bytes to Go slice
	goBytes := C.GoBytes(unsafe.Pointer(updateBytes), updateLen)
	result := make([]byte, len(goBytes))
	copy(result, goBytes)

	return result, nil
}

// applyUpdatesToYDoc applies Y.js updates to a base document using real YFFI
func applyUpdatesToYDoc(baseSnapshot []byte, updates [][]byte) ([]byte, error) {
	// Create a new Y.js document
	doc := C.ydoc_new()
	if doc == nil {
		return nil, fmt.Errorf("failed to create Y.js document")
	}
	defer C.ydoc_destroy(doc)

	// Apply base snapshot if provided
	if len(baseSnapshot) > 0 {
		txn := C.ydoc_write_transaction(doc, 0, nil)
		if txn == nil {
			return nil, fmt.Errorf("failed to create transaction for base snapshot")
		}

		// Apply base state
		success := C.ytransaction_apply(txn, (*C.uchar)(unsafe.Pointer(&baseSnapshot[0])), C.int(len(baseSnapshot)))
		C.ytransaction_commit(txn)

		if success != 0 {
			return nil, fmt.Errorf("failed to apply base snapshot: error code %d", success)
		}
	}

	// Apply each update
	for i, update := range updates {
		if len(update) == 0 {
			continue
		}

		txn := C.ydoc_write_transaction(doc, 0, nil)
		if txn == nil {
			return nil, fmt.Errorf("failed to create transaction for update %d", i)
		}

		// Apply update
		success := C.ytransaction_apply(txn, (*C.uchar)(unsafe.Pointer(&update[0])), C.int(len(update)))
		C.ytransaction_commit(txn)

		if success != 0 {
			return nil, fmt.Errorf("failed to apply update %d: error code %d", i, success)
		}
	}

	// Get final document state
	txn := C.ydoc_read_transaction(doc)
	if txn == nil {
		return nil, fmt.Errorf("failed to create read transaction")
	}
	defer C.ytransaction_commit(txn)

	var finalLen C.int
	finalBytes := C.ytransaction_state_diff_v1(txn, nil, 0, &finalLen)
	if finalBytes == nil {
		return nil, fmt.Errorf("failed to get final document state")
	}
	defer C.ybinary_destroy(finalBytes, finalLen)

	// Convert to Go bytes
	goBytes := C.GoBytes(unsafe.Pointer(finalBytes), finalLen)
	result := make([]byte, len(goBytes))
	copy(result, goBytes)

	return result, nil
}

// extractTextFromYjsDocument extracts text content from a Y.js document snapshot
func extractTextFromYjsDocument(snapshotData []byte, fieldName string) (string, error) {
	if len(snapshotData) == 0 {
		return "", fmt.Errorf("snapshot data cannot be empty")
	}

	if fieldName == "" {
		return "", fmt.Errorf("field name cannot be empty")
	}

	// Create a new Y.js document
	doc := C.ydoc_new()
	if doc == nil {
		return "", fmt.Errorf("failed to create Y.js document")
	}
	defer C.ydoc_destroy(doc)

	// Get the text field first (before applying snapshot)
	cFieldName := C.CString(fieldName)
	if cFieldName == nil {
		return "", fmt.Errorf("failed to create C string for field name")
	}
	defer C.free(unsafe.Pointer(cFieldName))

	textField := C.ytext(doc, cFieldName)
	if textField == nil {
		return "", fmt.Errorf("failed to get text field '%s'", fieldName)
	}

	// Start a write transaction to apply the snapshot and read text
	txn := C.ydoc_write_transaction(doc, 0, nil)
	if txn == nil {
		return "", fmt.Errorf("failed to create write transaction")
	}
	defer C.ytransaction_commit(txn)

	// Apply the snapshot data as an update
	updatePtr := (*C.uchar)(unsafe.Pointer(&snapshotData[0]))
	result := C.ytransaction_apply(txn, updatePtr, C.int(len(snapshotData)))
	if result != 0 {
		return "", fmt.Errorf("failed to apply snapshot data: error code %d", result)
	}

	// Extract text content from the field within the same transaction
	textStr := C.ytext_string(textField, txn)
	if textStr == nil {
		// Empty text is valid - return empty string
		return "", nil
	}
	defer C.ystring_destroy(textStr)

	// Convert C string to Go string
	return C.GoString(textStr), nil
}

// loadYDocFromData loads a YJS document from binary data
func loadYDocFromData(data []byte) (*C.YDoc, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("input data cannot be empty")
	}

	// Create a new Y.js document
	doc := C.ydoc_new()
	if doc == nil {
		return nil, fmt.Errorf("failed to create Y.js document")
	}

	// Start a write transaction to apply the data
	txn := C.ydoc_write_transaction(doc, 0, nil)
	if txn == nil {
		C.ydoc_destroy(doc)
		return nil, fmt.Errorf("failed to create write transaction")
	}

	// Apply the binary data to the document
	updatePtr := (*C.uchar)(unsafe.Pointer(&data[0]))
	result := C.ytransaction_apply(txn, updatePtr, C.int(len(data)))
	if result != 0 {
		C.ytransaction_commit(txn)
		C.ydoc_destroy(doc)
		return nil, fmt.Errorf("failed to apply data: error code %d", result)
	}
	C.ytransaction_commit(txn)

	return doc, nil
}

// destroyYDoc destroys a YJS document
func destroyYDoc(doc *C.YDoc) {
	if doc != nil {
		C.ydoc_destroy(doc)
	}
}

// getYMap gets a YMap from a document by name
func getYMap(doc *C.YDoc, name string) (*C.Branch, error) {
	if doc == nil {
		return nil, fmt.Errorf("document is nil")
	}

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	ymap := C.ymap(doc, cName)
	if ymap == nil {
		return nil, fmt.Errorf("failed to get map '%s'", name)
	}

	return ymap, nil
}

// createReadTransaction creates a read transaction for a document
func createReadTransaction(doc *C.YDoc) (*C.YTransaction, error) {
	if doc == nil {
		return nil, fmt.Errorf("document is nil")
	}

	txn := C.ydoc_read_transaction(doc)
	if txn == nil {
		return nil, fmt.Errorf("failed to create read transaction")
	}

	return txn, nil
}

// commitTransaction commits a transaction
func commitTransaction(txn *C.YTransaction) {
	if txn != nil {
		C.ytransaction_commit(txn)
	}
}

// getMapLength gets the number of entries in a map
func getMapLength(ymap *C.Branch, txn *C.YTransaction) (uint32, error) {
	if ymap == nil || txn == nil {
		return 0, fmt.Errorf("map or transaction is nil")
	}

	return uint32(C.ymap_len(ymap, txn)), nil
}

// createMapIterator creates an iterator for a map
func createMapIterator(ymap *C.Branch, txn *C.YTransaction) (*C.YMapIter, error) {
	if ymap == nil || txn == nil {
		return nil, fmt.Errorf("map or transaction is nil")
	}

	iter := C.ymap_iter(ymap, txn)
	if iter == nil {
		return nil, fmt.Errorf("failed to create map iterator")
	}

	return iter, nil
}

// destroyMapIterator destroys a map iterator
func destroyMapIterator(iter *C.YMapIter) {
	if iter != nil {
		C.ymap_iter_destroy(iter)
	}
}

// MapEntry represents a key-value pair from a YJS map
type MapEntry struct {
	Key   string
	Value *C.YOutput
}

// nextMapEntry gets the next entry from a map iterator
func nextMapEntry(iter *C.YMapIter) (*MapEntry, error) {
	if iter == nil {
		return nil, fmt.Errorf("iterator is nil")
	}

	entry := C.ymap_iter_next(iter)
	if entry == nil {
		return nil, nil // End of iteration
	}

	return &MapEntry{
		Key:   C.GoString(entry.key),
		Value: entry.value,
	}, nil
}

// extractContentFromYOutput extracts string content from a YOutput
func extractContentFromYOutput(output *C.YOutput, txn *C.YTransaction) (string, error) {
	if output == nil {
		return "", fmt.Errorf("output is nil")
	}

	// Try to read as string first
	strPtr := C.youtput_read_string(output)
	if strPtr != nil {
		content := C.GoString(strPtr)
		C.ystring_destroy(strPtr)
		return content, nil
	}

	// Try to read as YText
	textBranch := C.youtput_read_ytext(output)
	if textBranch != nil {
		textStr := C.ytext_string(textBranch, txn)
		if textStr != nil {
			content := C.GoString(textStr)
			C.ystring_destroy(textStr)
			return content, nil
		}
		return "", nil // Empty text
	}

	// Try to read as YDoc (nested document with file content)
	ydoc := C.youtput_read_ydoc(output)
	if ydoc != nil {
		// The file content is stored as a YDoc with a "content" text field
		contentFieldName := C.CString("content")
		defer C.free(unsafe.Pointer(contentFieldName))

		contentText := C.ytext(ydoc, contentFieldName)
		if contentText != nil {
			// Create a read transaction for the nested doc
			nestedTxn := C.ydoc_read_transaction(ydoc)
			if nestedTxn != nil {
				defer C.ytransaction_commit(nestedTxn)

				textStr := C.ytext_string(contentText, nestedTxn)
				if textStr != nil {
					content := C.GoString(textStr)
					C.ystring_destroy(textStr)
					return content, nil
				}
			}
		}
		return "", nil // Empty document
	}

	// Try to read as binary string
	binPtr := C.youtput_read_binary(output)
	if binPtr != nil {
		// Binary data as null-terminated string
		content := C.GoString(binPtr)
		return content, nil
	}

	// Try to read as nested YDoc (another map)
	nestedMap := C.youtput_read_ymap(output)
	if nestedMap != nil {
		// For nested documents, try to extract content from a "content" field
		contentKey := C.CString("content")
		defer C.free(unsafe.Pointer(contentKey))

		contentOutput := C.ymap_get(nestedMap, txn, contentKey)
		if contentOutput != nil {
			defer C.youtput_destroy(contentOutput)
			return extractContentFromYOutput(contentOutput, txn)
		}
		return "", fmt.Errorf("nested map found but no 'content' field")
	}

	return "", fmt.Errorf("unsupported YOutput type")
}

// yjsMapIteratorImpl implements YjsMapIterator
type yjsMapIteratorImpl struct {
	doc   *C.YDoc
	iter  *C.YMapIter
	txn   *C.YTransaction
	count uint32
}

// newYjsMapIterator creates a new map iterator
func newYjsMapIterator(data []byte, mapName string) (YjsMapIterator, error) {
	doc, err := loadYDocFromData(data)
	if err != nil {
		return nil, err
	}

	ymap, err := getYMap(doc, mapName)
	if err != nil {
		destroyYDoc(doc)
		return nil, err
	}

	txn, err := createReadTransaction(doc)
	if err != nil {
		destroyYDoc(doc)
		return nil, err
	}

	count, err := getMapLength(ymap, txn)
	if err != nil {
		commitTransaction(txn)
		destroyYDoc(doc)
		return nil, err
	}

	iter, err := createMapIterator(ymap, txn)
	if err != nil {
		commitTransaction(txn)
		destroyYDoc(doc)
		return nil, err
	}

	return &yjsMapIteratorImpl{
		doc:   doc,
		iter:  iter,
		txn:   txn,
		count: count,
	}, nil
}

func (it *yjsMapIteratorImpl) Next() (string, string, error) {
	if it.iter == nil || it.txn == nil {
		return "", "", fmt.Errorf("iterator is closed")
	}

	mapEntry, err := nextMapEntry(it.iter)
	if err != nil {
		return "", "", err
	}
	if mapEntry == nil {
		return "", "", nil // End of iteration
	}

	content, err := extractContentFromYOutput(mapEntry.Value, it.txn)
	if err != nil {
		return "", "", fmt.Errorf("failed to extract content for key '%s': %v", mapEntry.Key, err)
	}

	return mapEntry.Key, content, nil
}

func (it *yjsMapIteratorImpl) Count() uint32 {
	return it.count
}

func (it *yjsMapIteratorImpl) Close() {
	if it.iter != nil {
		destroyMapIterator(it.iter)
		it.iter = nil
	}
	if it.txn != nil {
		commitTransaction(it.txn)
		it.txn = nil
	}
	if it.doc != nil {
		destroyYDoc(it.doc)
		it.doc = nil
	}
}
