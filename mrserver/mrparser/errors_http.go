package mrparser

import (
	. "github.com/mondegor/go-sysmess/mrerr"
)

var (
	FactoryErrHttpRequestFileSizeMin = NewFactory(
		"errHttpRequestFileSizeMin", ErrorTypeUser, "invalid file size, min size = {{ .value }}b")

	FactoryErrHttpRequestFileSizeMax = NewFactory(
		"errHttpRequestFileSizeMax", ErrorTypeUser, "invalid file size, max size = {{ .value }}b")

	FactoryErrHttpRequestFileExtension = NewFactory(
		"errHttpRequestFileExtension", ErrorTypeUser, "invalid file extension: {{ .value }}")

	FactoryErrHttpRequestFileTotalSizeMax = NewFactory(
		"errHttpRequestFileTotalSizeMax", ErrorTypeUser, "invalid file total size, max total size = {{ .value }}b")

	FactoryErrHttpRequestFileContentType = NewFactory(
		"errHttpRequestFileContentType", ErrorTypeUser, "the content type '{{ .value }}' does not match the detected type")

	FactoryErrHttpRequestFileUnsupportedType = NewFactory(
		"errHttpRequestFileUnsupportedType", ErrorTypeUser, "unsupported file type '{{ .value }}'")

	FactoryErrHttpRequestImageWidthMax = NewFactory(
		"errHttpRequestImageWidthMax", ErrorTypeUser, "invalid image width, max size = {{ .value }}px")

	FactoryErrHttpRequestImageHeightMax = NewFactory(
		"errHttpRequestImageHeightMax", ErrorTypeUser, "invalid image height, max size = {{ .value }}px")
)

func WrapFileError(err error, name string) error {
	if FactoryErrHttpRequestFileSizeMin.Is(err) {
		return NewCustomError(name, err)
	}

	if FactoryErrHttpRequestFileSizeMax.Is(err) {
		return NewCustomError(name, err)
	}

	if FactoryErrHttpRequestFileExtension.Is(err) {
		return NewCustomError(name, err)
	}

	if FactoryErrHttpRequestFileContentType.Is(err) {
		return NewCustomError(name, err)
	}

	if FactoryErrHttpRequestFileUnsupportedType.Is(err) {
		return NewCustomError(name, err)
	}

	return err
}

func WrapImageError(err error, name string) error {
	if FactoryErrHttpRequestImageWidthMax.Is(err) {
		return NewCustomError(name, err)
	}

	if FactoryErrHttpRequestImageHeightMax.Is(err) {
		return NewCustomError(name, err)
	}

	return WrapFileError(err, name)
}
