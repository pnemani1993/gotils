package client

import (
	"fmt"
	"io"
	"net/http"
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
		return &InvalidOperation{1001, "Invalid method usage. See 'Addf' function"}
	}
	subPath, _ = strings.CutPrefix(subPath, "/")
	subPath, _ = strings.CutSuffix(subPath, "/")

	path.paths = append(path.paths, subPath)
	return nil
}

func (path *Path) Addf(subPath string, values []string) error {
	if strings.Count(subPath, "{") != strings.Count(subPath, "}") || strings.Count(subPath, "{") != len(values) {
		if strings.Count(subPath, "{") != strings.Count(subPath, "}") {
			return &InvalidInput{1001, "invalid path provided", subPath}
		} else {
			return &InvalidInput{1002, "all the required substitutions are not provided", subPath}
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

type InvalidOperation struct {
	code    int
	message string
}

func (err *InvalidOperation) Error() string {
	return fmt.Sprintf("Error %d: %s", err.code, err.message)
}

type InvalidInput struct {
	code    int
	message string
	input   string
}

func (err *InvalidInput) Error() string {
	return fmt.Sprintf("Error %d: %s\n\tInput: %s", err.code, err.message, err.input)
}

type HttpError struct {
	statusCode int
	message    string
}

func (err *HttpError) Error() string {
	return fmt.Sprintf("Error %d: %s", err.statusCode, err.message)
}

type ClientError struct {
	errorsList []error
}

func (err *ClientError) Error() string {
	errorString := ""
	for _, errs := range err.errorsList {
		errorString = errorString + errs.Error() + "\n"
	}
	return errorString
}

type HttpFuture struct {
	responseChannel chan int
	err             error
	isDone          bool
	response        *http.Response
}

func (future *HttpFuture) Get() (*http.Response, error) {

	<-future.responseChannel

	if future.err != nil {
		return future.response, future.err
	}

	if future.response.StatusCode > 299 {
		responseMessage, _ := io.ReadAll(future.response.Body)
		errorResponse := &HttpError{future.response.StatusCode, string(responseMessage)}
		return future.response, errorResponse
	}
	return future.response, nil
}

func (future *HttpFuture) IsDone() bool {
	return future.isDone
}
