package mrreq

import (
	"net/http"
	"strings"

	"github.com/mondegor/go-webcore/mrcore"
)

const (
	maxLenString = 256
)

func ParseStr(r *http.Request, key string, required bool) (string, error) {
	value := strings.TrimSpace(r.URL.Query().Get(key))

	if value == "" {
		if required {
			return "", mrcore.FactoryErrHTTPRequestParamEmpty.New(key)
		}

		return "", nil
	}

	if len(value) > maxLenString {
		return "", mrcore.FactoryErrHTTPRequestParamLenMax.New(key, maxLenString)
	}

	return value, nil
}
