package domainapp

import "path/filepath"

// removeXoXouid deletes all .xo-xouid.go files.
func (d *AppDomain) removeXoXouid(dir string) error {
	return d.deleteGlob(filepath.Join(dir, "*.xo-xouid.go"))
}
