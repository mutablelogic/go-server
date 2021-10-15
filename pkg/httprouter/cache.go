package httprouter

import (
	"math"
	"net/http"
	"sort"
	"sync"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type cache struct {
	sync.RWMutex

	cache map[string]*cachehandler
}

type cachehandler struct {
	hits    int64
	params  []string
	handler http.HandlerFunc
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	maxCacheSize = 100
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Get parameters and handler from cache
func (this *cache) Get(key string) (http.HandlerFunc, []string, int64) {
	this.init()

	// Prune cache in background
	if len(this.cache) > maxCacheSize {
		go this.clean()
	}

	// Check cache
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()
	if handler, exists := this.cache[key]; exists {
		if handler.hits < math.MaxInt64 {
			handler.hits++
		}
		return handler.handler, handler.params, handler.hits
	} else {
		return nil, nil, 0
	}
}

// Set parameters and handler into cache, in the background
func (this *cache) Set(key string, handler http.HandlerFunc, params []string) {
	this.init()
	go func() {
		this.RWMutex.Lock()
		defer this.RWMutex.Unlock()
		this.cache[key] = &cachehandler{
			hits:    0,
			params:  arrcopy(params),
			handler: handler,
		}
	}()
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Initialise the cache
func (this *cache) init() {
	if this.cache != nil {
		return
	}
	this.RWMutex.Lock()
	this.cache = make(map[string]*cachehandler)
	this.RWMutex.Unlock()
}

// Clean the cache
func (this *cache) clean() {
	arr := make([]int64, 0, len(this.cache))

	// Make sorted array of hits
	this.RWMutex.RLock()
	for _, handler := range this.cache {
		arr = append(arr, handler.hits)
	}
	this.RWMutex.RUnlock()

	// Return if no hits
	n := len(arr)
	if n == 0 {
		return
	}

	// Sort the hits array
	sort.Slice(arr, func(i, j int) bool {
		return arr[i] > arr[j]
	})

	// Lock for cleaning the cache
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()

	// Remove all entries which are less than or equal to the median
	// and reduce the hit counter for the remainder
	median := arr[(n-1)/2]
	for key, handler := range this.cache {
		if handler.hits <= median {
			delete(this.cache, key)
		} else if handler.hits > 0 {
			handler.hits = handler.hits - 1
		}
	}
}

// Copy array
func arrcopy(a []string) []string {
	b := make([]string, len(a))
	for i := range a {
		b[i] = a[i]
	}
	return b
}
