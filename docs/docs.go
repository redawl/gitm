// Package docs holds the docs repository implementation,
// for reading bundled *.md files
package docs

import (
	"embed"
	"io/fs"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage/repository"
)

//go:embed **.md
var docs embed.FS

var _ repository.Repository = (*DocsRepository)(nil)

type DocsRepository struct{}

type docsFile struct {
	fs.File
	uri fyne.URI
}

func (d *docsFile) URI() fyne.URI {
	return d.uri
}

func (d *DocsRepository) Exists(u fyne.URI) (bool, error) {
	_, err := docs.Open(u.Path())
	if err != nil {
		return false, nil
	}

	return true, nil
}

func (d *DocsRepository) Reader(u fyne.URI) (fyne.URIReadCloser, error) {
	file, err := docs.Open(u.Path())
	if err != nil {
		return nil, err
	}

	return &docsFile{File: file, uri: u}, nil
}

func (d *DocsRepository) CanRead(u fyne.URI) (bool, error) {
	return d.Exists(u)
}

func (d *DocsRepository) Destroy(string) {}
