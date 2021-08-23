package httprouter

import (
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
	hits    uint
	params  []string
	handler *http.HandlerFunc
}

type cachefreq struct {
	hits uint
	key  string
}

type cachefreqarr []cachefreq

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	maxCacheSize = 1000
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Get parameters and handler from cache
func (this *cache) Get(key string) (http.HandlerFunc, []string, uint) {
	this.init()
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()

	// Prune cache in background
	if len(this.cache) > maxCacheSize {
		go this.clean()
	}

	// Check cache
	if handler, exists := this.cache[key]; exists {
		handler.hits++
		return *handler.handler, handler.params, handler.hits
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
			params:  arrcopy(params),
			handler: &handler,
		}
	}()
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Initialise the cache
func (this *cache) init() {
	if this.cache == nil {
		this.RWMutex.Lock()
		defer this.RWMutex.Unlock()
		this.cache = make(map[string]*cachehandler)
	}
}

// Clean the cache
func (this *cache) clean() {
	// Make sorted array of cache entries
	arr := cachefreqarr{}
	this.RWMutex.RLock()
	for key, handler := range this.cache {
		arr = append(arr, cachefreq{hits: handler.hits, key: key})
	}
	sort.Sort(arr)
	this.RWMutex.RUnlock()

	// Lock for cleaning the cache
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()

	// Now remove entries
	for i := (maxCacheSize << 1); i < len(arr); i++ {
		delete(this.cache, arr[i].key)
	}

	// Remove one hit from each entry
	for key, handler := range this.cache {
		if handler.hits > 0 {
			handler.hits--
		} else {
			delete(this.cache, key)
		}
	}
}

// Sort array
func (this cachefreqarr) Len() int {
	return len(this)
}

func (this cachefreqarr) Less(i, j int) bool {
	return this[i].hits > this[j].hits
}

func (this cachefreqarr) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

// Copy array
func arrcopy(a []string) []string {
	b := make([]string, len(a))
	for i := range a {
		b[i] = a[i]
	}
	return b
}
