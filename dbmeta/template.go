package dbmeta

import (
	"bytes"
	"go/format"
	"text/template"
)

func RenderModelTemplate(t *template.Template, m *ModelInfo) ([]byte, error) {
	// body, err := Asset(filepath.Join(templateDir, "controller.go.tmpl"))

	// if err != nil {
	// 	return nil, err
	// }

	// tmpl, err := template.New("controller").Funcs(funcMap).Parse(string(body))

	// if err != nil {
	// 	return err
	// }

	var buf bytes.Buffer
	if err := t.Execute(&buf, m); err != nil {
		return nil, err
	}
	src, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, err
	}

	return src, nil
}
