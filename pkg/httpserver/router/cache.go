package router

/* TODO
type cached struct {
	index   int
	matched []string
}

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewWithConfig(c Config) (Router, error) {
	r := new(router)
	r.cache = make(map[string]*cached)

	// Return success
	return r, nil
}

/////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r *router) String() string {
	str := "<router"
	for _, route := range r.routes {
		str += fmt.Sprintf(" %q %q => %q", route.prefix, route.path, route.methods)
	}
	return str + ">"
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// AddHandler adds a handler to the router, for a specific prefix and http methods supported.
// If the path argument is nil, then any path under the prefix will match. If the path contains
// a regular expression, then a match is made and any matched parameters of the regular
// expression can be retrieved from the request context.
func (r *router) AddHandler(gateway Gateway, path *regexp.Regexp, fn http.HandlerFunc, methods ...string) error {
	// Check gateway
	if gateway == nil {
		return ErrBadParameter.With("gateway")
	}

	// If methods is empty, default to GET
	if len(methods) == 0 {
		methods = []string{"GET"}
	}

	// Append the route
	r.routes = append(r.routes, route{normalizePath(gateway.Prefix(), true), path, fn, methods})

	// Sort routes by prefix length, longest first, and then by path != nil vs nil
	sort.Slice(r.routes, func(i, j int) bool {
		if len(r.routes[i].prefix) < len(r.routes[j].prefix) {
			return false
		}
		if len(r.routes[i].prefix) == len(r.routes[j].prefix) && r.routes[i].path == nil {
			return false
		}
		return true
	})

	// Return success
	return nil
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	route, params := r.get(req.Method, req.URL.Path)
	if route == nil {
		util.ServeError(w, http.StatusNotFound)
		return
	}

	// Check methods
	if slices.Contains(route.methods, req.Method) {
		route.fn(w, req.Clone(context.WithPrefixParams(req.Context(), route.prefix, params)))
		return
	}

	// Return method not allowed
	util.ServeError(w, http.StatusMethodNotAllowed)
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// get returns the route for the given path and method, and the parameters matched
// or returns nil for the route otherwise
func (r *router) get(method, path string) (*route, []string) {
	// Check cache
	if route, params := r.getcached(method, path); route != nil {
		return route, params
	}

	// Search routes to find candidates
	methodAllowed := true
	for i := range r.routes {
		route := r.routes[i]

		// Check against the prefix
		if !strings.HasPrefix(path, route.prefix) {
			continue
		}

		// Check for default route: this is the route that matches everything
		if route.path == nil {
			if contains(route.methods, method) {
				r.setcached(method, path, i, nil)
				return &route, nil
			}
			methodAllowed = false
			continue
		}

		// Check with a regular expression
		relpath := normalizePath(path[len(route.prefix):], false)
		if params := route.path.FindStringSubmatch(relpath); params != nil {
			if contains(route.methods, method) {
				r.setcached(method, path, i, params[1:])
				return &route, nil
			}
			methodAllowed = false
			continue
		}
	}

	if !methodAllowed {
		fmt.Println("TODO: methodNotAllowed", method, path)
	}

	// No match
	return nil, nil
}

// getcached returns the route for the given path, and the parameters matched
// or returns nil for the route otherwise
func (r *router) getcached(method, path string) (*route, []string) {
	r.RLock()
	defer r.RUnlock()
	cached, exists := r.cache[method+path]
	if !exists {
		return nil, nil
	} else {
		return &r.routes[cached.index], cached.matched
	}
}

// setcached puts a route into the cache
func (r *router) setcached(method, path string, index int, params []string) {
	r.Lock()
	defer r.Unlock()
	r.cache[method+path] = &cached{index, params}
}

// contains returns true if a string array contains a string
func contains(a []string, s string) bool {
	return slices.Contains(a, s)
}
*/
