package template

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/oxtoacart/bpool"

	"github.com/funnydog/mailadmin/core/config"
	"github.com/funnydog/mailadmin/core/urls"
)

type Manager struct {
	bufferPool *bpool.BufferPool
	templates  map[string]*template.Template
}

func (m *Manager) Render(w http.ResponseWriter, base, name string, data *map[string]interface{}) error {
	tmp, ok := m.templates[name]
	if !ok {
		return fmt.Errorf("Template %s not found", name)
	}

	buf := m.bufferPool.Get()
	defer m.bufferPool.Put(buf)

	var err error
	if base != "" {
		err = tmp.ExecuteTemplate(buf, base, data)
	} else {
		err = tmp.Execute(buf, data)
	}
	if err != nil {
		return err
	}

	_, err = buf.WriteTo(w)
	return err
}

func Create(fsys fs.FS, conf config.Static, um *urls.Manager) (Manager, error) {

	// generic tags
	tags, err := fs.Glob(fsys, filepath.Join(conf.TagsDir, "*.html"))
	if err != nil {
		tags = []string{}
	}

	// templates
	filenames, err := fs.Glob(fsys, filepath.Join(conf.TemplateDir, "*.html"))
	if err != nil {
		return Manager{}, err
	}

	// layouts extended by templates
	extends, err := fs.Glob(fsys, filepath.Join(conf.ExtendDir, "*.html"))
	if err != nil {
		return Manager{}, err
	}

	fmap := template.FuncMap{
		"reverse": func(url string, args ...interface{}) (string, error) {
			return um.Reverse(url, args)
		},
		"static": func(path string) string {
			return um.Static(path)
		},
	}

	templates := map[string]*template.Template{}
	for _, file := range filenames {

		name := filepath.Base(file)
		t := template.New(name).Funcs(fmap)

		files := append(extends, tags...)
		files = append(files, file)
		templates[name] = template.Must(t.ParseFS(fsys, files...))
	}

	bp := bpool.NewBufferPool(64)

	return Manager{
		bufferPool: bp,
		templates:  templates,
	}, nil
}
