package client

import (
	"testing"
)

func TestAdd(t *testing.T) {
	path := NewPath()

	path.Add("/path1/path2/path3/")
	path.Add("path4/path5/")
	path.Add("/path6/path7")

	expected1 := "path1/path2/path3"
	expected2 := "path4/path5"
	expected3 := "path6/path7"

	if expected1 != path.paths[0] {
		t.Fatalf("expected: %s;\n actual: %s", expected1, path.paths[0])
	}

	if expected2 != path.paths[1] {
		t.Fatalf("expected: %s;\n actual: %s", expected2, path.paths[1])
	}

	if expected3 != path.paths[2] {
		t.Fatalf("expected: %s;\n actual: %s", expected3, path.paths[2])
	}
}

func TestAddInvalid(t *testing.T) {
	path := NewPath()
	err := path.Add("/path1/{path2}/path3/")
	if err == nil {
		t.Fatalf("expected error to be thrown")
	}

	err = path.Add("/path1/path2/path3/")
	if err != nil {
		t.Fatalf("expected no error to be thrown")
	}
}

func TestAddf(t *testing.T) {
	path := NewPath()
	path.Addf("/path1/{path2}/path3/", "value")
	expected := "path1/value/path3"
	actual := path.paths[0]

	if actual != expected {
		t.Fatalf("expected: %s;\n actual: %s", expected, actual)
	}
}

func TestGetPath(t *testing.T) {
	path := NewPath()

	path.Add("/path1/path2/path3/")
	path.Add("path4/path5/")
	path.Add("/path6/path7")

	expected := "/path1/path2/path3/path4/path5/path6/path7"

	if expected != path.GetPath() {
		t.Fatalf("expected: %s;\n actual: %s", expected, path.GetPath())
	}
}

func TestPatternSubstituter(t *testing.T) {
	path := "/path1/{someValue}/path2"
	someValue := "value"
	expected := "/path1/value/path2"

	result := patternSubstituter(path, someValue)
	if expected != result {
		t.Fatalf("expected: %s;\n actual: %s", expected, result)
	}
}
