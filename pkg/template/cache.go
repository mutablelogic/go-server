package template

import (
	"html/template"
	"io"
	"io/fs"
	"sort"
	"strconv"
	"sync"

	// Modules
	. "github.com/djthorpe/go-server"
	multierror "github.com/hashicorp/go-multierror"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Cache struct {
	sync.RWMutex
	fs.FS
	*template.Template
	infos map[string]fs.FileInfo
	funcs template.FuncMap
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	DefaultPattern = "*.tmpl"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func NewCache(dir fs.FS) (*Cache, error) {
	this := new(Cache)
	this.FS = dir

	// Read templates
	if err := this.read(); err != nil {
		return nil, err
	}

	// Return success
	return this, nil
}

func (this *Cache) Lookup(name string) (*template.Template, error) {
	// Lookup existing template
	this.RWMutex.RLock()
	info, exists := this.infos[name]
	this.RWMutex.RUnlock()
	if exists == false {
		return nil, ErrNotFound.With(name)
	}

	// Get stat on templates, might re-load template from disk
	file, err := this.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Compare template ModTime and reload templates as necessary
	var result error
	if info2, err := file.Stat(); err == nil {
		if !info2.ModTime().Equal(info.ModTime()) {
			if err := this.read(); err != nil {
				result = multierror.Append(result, err)
			}
		}
	}

	// Return template
	tmpl := this.Template.Lookup(info.Name())
	if tmpl == nil {
		result = multierror.Append(result, ErrBadParameter.With(name))
	}
	return tmpl, result
}

func (this *Cache) Exec(w io.Writer, name string, v interface{}) error {
	if tmpl, err := this.Lookup(name); err != nil {
		return err
	} else if err := tmpl.Execute(w, v); err != nil {
		return err
	}
	// Return success
	return nil
}

func (this *Cache) Templates() []string {
	var result []string
	for _, template := range this.Template.Templates() {
		if name := template.Name(); name != "" {
			result = append(result, name)
		}
	}
	return result
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Cache) String() string {
	str := "<cache templates=["
	tmpls := ""
	for _, template := range this.Template.Templates() {
		if tmpls != "" {
			tmpls += ", "
		}
		if name := template.Name(); name != "" {
			tmpls += strconv.Quote(template.Name())
		}
	}
	return str + tmpls + "]>"
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *Cache) read() error {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()

	// Get list of files to be used as templates
	names, err := fs.Glob(this.FS, DefaultPattern)
	if err != nil {
		return err
	} else if len(names) == 0 {
		return ErrNotFound.With("No templates found")
	}

	// Sort names in alphabetical order
	sort.Strings(names)

	// Extract modtime
	infos := make(map[string]fs.FileInfo)
	for _, name := range names {
		file, err := this.FS.Open(name)
		if err != nil {
			return err
		}
		defer file.Close()

		info, err := file.Stat()
		if err != nil {
			return err
		} else {
			infos[name] = info
		}
	}

	// Read all templates
	tmpl, err := template.New("").Funcs(this.funcs).ParseFS(this.FS, names...)
	if err != nil {
		return err
	}

	// Set template cache
	this.infos = infos
	this.Template = tmpl

	// Return success
	return nil
}
