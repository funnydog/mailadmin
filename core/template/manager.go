package template

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/oxtoacart/bpool"

	"github.com/funnydog/mailadmin/core/urls"
)

type Manager struct {
	bufferPool *bpool.BufferPool
	basename   string
	templates  map[string]*template.Template
}

func (m *Manager) Render(w http.ResponseWriter, name string, data *map[string]interface{}) error {
	tmp, ok := m.templates[name]
	if !ok {
		return fmt.Errorf("Template %s not found", name)
	}

	buf := m.bufferPool.Get()
	defer m.bufferPool.Put(buf)

	err := tmp.ExecuteTemplate(buf, m.basename, data)
	if err != nil {
		return err
	}

	_, err = buf.WriteTo(w)
	return err
}

func Create(basename string, templateDir string, tagsDir string, um *urls.Manager) (Manager, error) {
	tags, err := filepath.Glob(filepath.Join(tagsDir, "*.html"))
	if err != nil {
		tags = []string{}
	}

	filenames, err := filepath.Glob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		return Manager{}, err
	}

	fmap := template.FuncMap{
		"reverse": func(url string, args ...interface{}) (string, error) {
			return um.Reverse(url, args)
		},
	}

	basefile := filepath.Join(templateDir, basename+".html")
	templates := map[string]*template.Template{}
	for _, file := range filenames {
		if file == basefile {
			continue
		}

		name := filepath.Base(file)
		t := template.New(name).Funcs(fmap)

		files := []string{
			basefile,
		}
		files = append(files, tags...)
		files = append(files, file)
		templates[name] = template.Must(t.ParseFiles(files...))
	}

	bp := bpool.NewBufferPool(64)

	return Manager{
		bufferPool: bp,
		basename:   basename,
		templates:  templates,
	}, nil
}
