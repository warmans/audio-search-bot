package docs

import (
	"embed"
	"fmt"
	"path"
	"slices"
	"strings"
)

//go:embed topics
var topics embed.FS

func NewRepo() (*Repo, error) {
	files, err := topics.ReadDir("topics")
	if err != nil {
		return nil, err
	}
	repo := &Repo{
		docs: make(map[string]string),
	}
	for _, f := range files {
		data, err := topics.ReadFile(path.Join("topics", f.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read embedded file %s: %w", f.Name(), err)
		}
		repo.docs[strings.TrimSuffix(path.Base(f.Name()), ".md")] = string(data)
	}
	return repo, nil
}

type Repo struct {
	docs map[string]string
}

func (r *Repo) Topics() []string {
	t := []string{}
	for name := range r.docs {
		t = append(t, name)
	}
	slices.Sort(t)
	return t
}

func (r *Repo) Get(name string) (string, error) {
	doc, ok := r.docs[name]
	if !ok {
		return "", fmt.Errorf("topic not found: %s", name)
	}
	return doc, nil
}
