package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// defaults is a persistent key-value store backed by a JSON file on disk.
type defaults struct {
	mu   sync.RWMutex
	path string
	data map[string]any
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// initialize a defaults store at the given file path.
// If the file exists, its contents are loaded; otherwise the store starts empty.
func (d *defaults) init(path string) error {
	d.path = path
	d.data = make(map[string]any)

	// Load existing file (ignore if it doesn't exist)
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	defer f.Close()

	// Read the JSON defaults file into the store.
	// If the file is corrupt or can't be read, discard
	if err := json.NewDecoder(f).Decode(&d.data); err != nil {
		d.data = make(map[string]any)
		if closeErr := f.Close(); closeErr != nil {
			return closeErr
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Get retrieves a value by key. Returns nil if the key does not exist.
func (d *defaults) Get(key string) any {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.data[key]
}

// GetString retrieves a string value by key. Returns empty string if the key
// does not exist or the value is not a string.
func (d *defaults) GetString(key string) string {
	v, _ := d.Get(key).(string)
	return v
}

// Set stores a value by key and persists the store to disk.
// Pass nil to remove a key.
func (d *defaults) Set(key string, value any) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.data == nil {
		d.data = make(map[string]any)
	}
	if value == nil {
		delete(d.data, key)
	} else {
		d.data[key] = value
	}
	return d.save()
}

// Keys returns all keys in the store.
func (d *defaults) Keys() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	keys := make([]string, 0, len(d.data))
	for k := range d.data {
		keys = append(keys, k)
	}
	return keys
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// save writes the store to disk as indented JSON, creating parent directories
// as needed.
func (d *defaults) save() error {
	if err := os.MkdirAll(filepath.Dir(d.path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(d.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(d.path, data, 0600)
}
