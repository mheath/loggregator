package dopplerproxy

import (
	"errors"
	"fmt"
	"net/http"
)

var MissingAppIdError = errors.New("No App Id specified in request")

func TranslateFromLegacyPath(request *http.Request) (*http.Request, error) {
	err := request.ParseForm()
	if err != nil {
		return nil, err
	}

	appId := request.Form.Get("app")

	if appId == "" {
		return nil, MissingAppIdError
	}

	translatedRequest := *request
	copiedUrl := *request.URL
	translatedRequest.URL = &copiedUrl

	switch request.URL.Path {

	case "/tail/":
		translatedRequest.URL.Path = fmt.Sprintf("/apps/%s/stream", appId)

	case "/dump/":
		fallthrough
	case "/recent":
		translatedRequest.URL.Path = fmt.Sprintf("/apps/%s/recentlogs", appId)

	default:
		return nil, fmt.Errorf("unexpected path: %s", request.URL.Path)
	}

	return &translatedRequest, nil
}

func TranslateFromDropsondePath(request *http.Request) (*http.Request, error) {
	return request, nil
}
