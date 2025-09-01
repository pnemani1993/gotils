package client

import (
	"fmt"
	"strings"
)

type Path struct {
	paths []string
}

func NewPath() Path {
	return Path{paths: make([]string, 0, 10)}
}

func (path *Path) Add(subPath string) error {
	if strings.ContainsAny(subPath, "{") && strings.ContainsAny(subPath, "}") {
		return fmt.Errorf("Invalid method usage. See 'Addf' function")
	}
	subPath, _ = strings.CutPrefix(subPath, "/")
	subPath, _ = strings.CutSuffix(subPath, "/")

	path.paths = append(path.paths, subPath)
	return nil
}

func (path *Path) Addf(subPath string, values ...string) error {
	if strings.Count(subPath, "{") != strings.Count(subPath, "}") || strings.Count(subPath, "{") != len(values) {
		if strings.Count(subPath, "{") != strings.Count(subPath, "}") {
			return fmt.Errorf("Invalid path provided: %s", subPath)
		} else {
			return fmt.Errorf("All the substitutions are not provided: %s", values)
		}
	}

	subPath, _ = strings.CutPrefix(subPath, "/")
	subPath, _ = strings.CutSuffix(subPath, "/")

	for _, subString := range values {
		subPath = patternSubstituter(subPath, subString)
	}

	path.paths = append(path.paths, subPath)
	return nil
}

func (path Path) GetPath() string {
	pathString := ""
	for _, value := range path.paths {
		pathString = pathString + "/" + value
	}
	return pathString
}

func patternSubstituter(str string, subString string) string {

	startIndex := strings.Index(str, "{")
	endIndex := strings.Index(str, "}")

	returnString := str[0:startIndex] + subString + str[endIndex+1:]
	return returnString
}
