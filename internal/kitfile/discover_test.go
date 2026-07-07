package kitfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscover(t *testing.T) {
	root := t.TempDir()

	write := func(rel, content string) {
		t.Helper()
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// recognized
	write(".kit", "")
	write("testes.kit", "")
	write(".kitfiles/.kit", "")
	write(".kitfiles/deploy", "")
	write(".kitfiles/db.yaml", "")
	write(".kitfiles/infra.xml", "")
	// ignored
	write("go.mod", "")
	write("README.md", "")
	write("config.yaml", "") // yaml outside .kitfiles is not a task file
	write(".kitfiles/notes.txt", "")
	write(".kitfiles/sub/x.kit", "") // no recursion inside .kitfiles

	files, err := Discover(root)
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]File{
		".kit":                {Scope: "", Format: FormatDSL},
		"testes.kit":          {Scope: "testes", Format: FormatDSL},
		".kitfiles/.kit":      {Scope: "", Format: FormatDSL},
		".kitfiles/deploy":    {Scope: "deploy", Format: FormatDSL},
		".kitfiles/db.yaml":   {Scope: "db", Format: FormatYAML},
		".kitfiles/infra.xml": {Scope: "infra", Format: FormatXML},
	}

	got := map[string]File{}
	for _, f := range files {
		rel, err := filepath.Rel(root, f.Path)
		if err != nil {
			t.Fatal(err)
		}
		got[filepath.ToSlash(rel)] = File{Scope: f.Scope, Format: f.Format}
	}

	for rel, w := range want {
		g, ok := got[rel]
		if !ok {
			t.Errorf("missing file %s", rel)
			continue
		}
		if g.Scope != w.Scope || g.Format != w.Format {
			t.Errorf("%s: got scope=%q format=%v, want scope=%q format=%v",
				rel, g.Scope, g.Format, w.Scope, w.Format)
		}
	}
	for rel := range got {
		if _, ok := want[rel]; !ok {
			t.Errorf("unexpected file discovered: %s", rel)
		}
	}
}

func TestDiscoverNoKitfilesDir(t *testing.T) {
	root := t.TempDir()
	files, err := Discover(root)
	if err != nil {
		t.Fatalf("missing .kitfiles dir should not fail: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected no files, got %d", len(files))
	}
}
