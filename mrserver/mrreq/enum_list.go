package mrreq

import (
	"strings"

	"github.com/mondegor/go-webcore/mrcore"
)

const (
	maxLenEnumList = 256
)

// ParseEnumList - возвращает массив Enum значений из внешнего строкового параметра по указанному ключу.
// Если параметр пустой, то возвращается пустой массив.
func ParseEnumList(getter valueGetter, key string) ([]string, error) {
	value := strings.TrimSpace(getter.Get(key))

	if value == "" {
		return nil, nil
	}

	if len(value) > maxLenEnumList {
		return nil, mrcore.ErrHttpRequestParamLenMax.New(key, maxLenEnumList)
	}

	items := strings.Split(strings.ToUpper(value), ",")

	for i, item := range items {
		item = strings.TrimSpace(item)

		if !regexpEnum.MatchString(item) {
			return nil, mrcore.ErrHttpRequestParseParam.New(key, "Enum", value)
		}

		items[i] = item
	}

	return items, nil
}
