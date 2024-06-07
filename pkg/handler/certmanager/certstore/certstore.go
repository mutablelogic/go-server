package certstore

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	// Packages
	"github.com/mutablelogic/go-server/pkg/handler/certmanager"
	"github.com/mutablelogic/go-server/pkg/handler/certmanager/cert"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////
// TYPES

// CertStore represents a certificate store
type certstore struct {
	// Root of the file storage
	dataPath string

	// Group for certificates on the file system
	fileGroup int

	// Mode for files
	fileMode os.FileMode
}

// Check interfaces are satisfied
var _ certmanager.CertStorage = (*certstore)(nil)

////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(c Config) (*certstore, error) {
	task := new(certstore)

	// Make data directory, set permissions
	if c.DataPath == "" {
		return nil, ErrBadParameter.With("missing 'data'")
	} else if gid, err := c.GroupId(); err != nil {
		return nil, err
	} else if err := os.MkdirAll(c.DataPath, c.DirMode()); err != nil {
		return nil, err
	} else if err := os.Chown(c.DataPath, -1, gid); err != nil {
		return nil, err
	} else if err := isWritableFileAtPath(c.DataPath); err != nil {
		return nil, ErrBadParameter.With("not writable: ", c.DataPath)
	} else {
		task.dataPath = c.DataPath
		task.fileGroup = gid
		task.fileMode = c.FileMode()
	}

	// Return success
	return task, nil
}

////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return all certificates
func (c *certstore) List() ([]certmanager.Cert, error) {
	// Read entries, silently ignore errors
	entries, err := os.ReadDir(c.dataPath)
	if err != nil {
		return nil, err
	}

	var result error
	certs := make([]certmanager.Cert, 0, len(entries))
	for _, entry := range entries {
		// Check for valid certificates
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != certExt {
			continue
		}
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if isReadableFileAtPath(filepath.Join(c.dataPath, entry.Name())) != nil {
			continue
		}

		// Read certificates and accumulate any errors
		serial := strings.TrimSuffix(entry.Name(), certExt)
		if cert, err := c.Read(strings.TrimSuffix(entry.Name(), certExt)); err != nil {
			result = errors.Join(result, fmt.Errorf("%q: %w", serial, err))
		} else if cert.Serial() != serial {
			result = errors.Join(result, ErrBadParameter.With("serial mismatch"))
		} else {
			certs = append(certs, cert)
		}
	}

	// Return certificates and any errors
	return certs, result
}

// Read a certificate
func (c *certstore) Read(serial string) (certmanager.Cert, error) {
	// Check for file
	pathForCert, err := c.pathForKey(serial)
	if err != nil {
		return nil, err
	}

	// Open file
	fh, err := os.Open(pathForCert)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	// Read data
	data, err := io.ReadAll(fh)
	if err != nil {
		return nil, err
	}

	// Parse certificate
	return cert.NewFromBytes(data)
}

// Create a new certificate
func (c *certstore) Write(cert certmanager.Cert) error {
	pathForCert, err := c.pathForKey(cert.Serial())
	if !errors.Is(err, os.ErrNotExist) {
		if err == nil {
			return ErrDuplicateEntry.With(cert.Serial())
		} else {
			return err
		}
	}

	// Create the certificate
	fh, err := os.Create(pathForCert)
	if err != nil {
		return err
	}
	defer fh.Close()

	// Write the certificate, set mode and group
	if err := cert.WriteCertificate(fh); err != nil {
		return errors.Join(err, os.Remove(pathForCert))
	} else if err := cert.WritePrivateKey(fh); err != nil {
		return errors.Join(err, os.Remove(pathForCert))
	} else if err := os.Chown(pathForCert, -1, c.fileGroup); err != nil {
		return errors.Join(err, os.Remove(pathForCert))
	} else if err := os.Chmod(pathForCert, c.fileMode); err != nil {
		return errors.Join(err, os.Remove(pathForCert))
	}

	// Return success
	return nil
}

// Delete a certificate
func (c *certstore) Delete(cert certmanager.Cert) error {
	pathForCert, err := c.pathForKey(cert.Serial())

	// Silently ignore "not exist" errors
	if errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}

	return os.Remove(pathForCert)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Returns the path for a certificate, and a boolean value which indicates
// if the folder exists
func (c *certstore) pathForKey(serial string) (string, error) {
	path := filepath.Join(c.dataPath, serial+certExt)
	if info, err := os.Stat(path); err != nil {
		return path, err
	} else if !info.Mode().IsRegular() {
		return path, ErrBadParameter.With("not a regular file: ", path)
	} else {
		return path, nil
	}
}

// Returns boolean value which indicates if a file is readable by current
// user
func isReadableFileAtPath(path string) error {
	return syscall.Access(path, R_OK)
}

// Returns boolean value which indicates if a file is writable by current
// user
func isWritableFileAtPath(path string) error {
	return syscall.Access(path, W_OK)
}

// Returns boolean value which indicates if a file is executable by current
// user
func isExecutableFileAtPath(path string) error {
	return syscall.Access(path, X_OK)
}
