package kitfile

import (
	"strings"
	"testing"
)

func TestParseDSL(t *testing.T) {
	src := `# top comment

build: compile the project
    go build ./...

nodesc:
    echo one
    echo two

release: multi-line with structure

    for os in linux windows; do
        echo "$os"
    done

    echo done
`
	tasks, err := ParseDSL([]byte(src), "test.kit")
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 3 {
		t.Fatalf("want 3 tasks, got %d: %+v", len(tasks), tasks)
	}

	build := tasks[0]
	if build.Name != "build" || build.Desc != "compile the project" || build.Run != "go build ./..." {
		t.Errorf("build parsed wrong: %+v", build)
	}

	nodesc := tasks[1]
	if nodesc.Desc != "" || nodesc.Run != "echo one\necho two" {
		t.Errorf("nodesc parsed wrong: %+v", nodesc)
	}

	release := tasks[2]
	want := "for os in linux windows; do\n    echo \"$os\"\ndone\n\necho done"
	if release.Run != want {
		t.Errorf("release body:\n%q\nwant:\n%q", release.Run, want)
	}
}

func TestParseDSLCRLF(t *testing.T) {
	src := "hi: says hi\r\n    echo hi\r\n"
	tasks, err := ParseDSL([]byte(src), "test.kit")
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].Run != "echo hi" {
		t.Fatalf("CRLF parse wrong: %+v", tasks)
	}
}

func TestParseDSLErrors(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string // substring expected in the error
	}{
		{"indented without task", "    echo hi\n", "indented line outside a task"},
		{"garbage at column 0", "build compile stuff\n", "expected `name: description`"},
		{"invalid name", "foo bar: desc\n    echo hi\n", "expected `name: description`"},
		{"duplicate task", "a: one\n    echo 1\na: two\n    echo 2\n", "already defined"},
		{"task without body", "a: empty\n\nb: ok\n    echo hi\n", "has no commands"},
		{"trailing empty task", "a: ok\n    echo hi\nb: empty\n", "has no commands"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseDSL([]byte(tc.src), "test.kit")
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %q does not contain %q", err, tc.want)
			}
		})
	}
}
