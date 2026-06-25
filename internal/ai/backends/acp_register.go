package backends

// ACP mapping data is now registered directly in each backend sub-package's
// init() via backends.Register(&BackendPlugin{...}).
// This file previously held centralized ACP registration via registerACP(),
// which caused import cycles when sub-packages imported the backends package.
// The registerACP() function is removed; all ACP data flows through Register().

// genericACPRemaps is the default 6-field camelCase→snake_case normalization map.
// Used as fallback when a backend has no specific ACP InputRemaps registered.
var genericACPRemaps = map[string]string{
	"oldString": "old_string", "newString": "new_string",
	"dirPath": "path", "filePath": "file_path",
	"cellIndex": "cell_index", "cellType": "cell_type",
}
