package mrmsg

import (
	"io"
	"strings"

	"github.com/mondegor/go-core/util/conv"
)

const (
	missingArg = "!MISSINGARG"
)

type (
	// PlaceholderReplacer - заменяет именованные плейсхолдеры в сообщении на аргументы
	// по порядку через strings.Replacer.
	PlaceholderReplacer struct {
		message      string
		placeholders []string
	}
)

// NewPlaceholderReplacer - создаёт PlaceholderReplacer для замены плейсхолдеров.
// Параметры:
//   - message - сообщение с именованными плейсхолдерами (например: "Hello {name}, you are {age}");
//   - placeholders - уникальные плейсхолдеры вместе с ограничителями в порядке аргументов.
func NewPlaceholderReplacer(message string, placeholders []string) *PlaceholderReplacer {
	return &PlaceholderReplacer{
		message:      message,
		placeholders: placeholders,
	}
}

// NewMessageReplacer - создаёт PlaceholderReplacer, автоматически извлекая плейсхолдеры из сообщения.
// Параметры:
//   - leftDelim, rightDelim - ограничители плейсхолдеров (например: "{" и "}");
//   - message - сообщение с плейсхолдерами для последующей замены.
func NewMessageReplacer(leftDelim, rightDelim, message string) *PlaceholderReplacer {
	return NewPlaceholderReplacer(
		message,
		NewPlaceholderExtractor(leftDelim, rightDelim).Extract(message),
	)
}

// Replace - подставляет аргументы в подготовленное сообщение, заменяя плейсхолдеры.
// Аргументы преобразуются в строки через conv.String и сопоставляются плейсхолдерам по порядку.
// Если плейсхолдеров нет, сообщение возвращается как есть.
// Недостающие аргументы заменяются на "!MISSINGARG", лишние игнорируются.
func (p *PlaceholderReplacer) Replace(args []any) (replacedMessage string, err error) {
	var buf strings.Builder

	if err = p.ReplaceTo(&buf, args); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ReplaceTo - записывает сообщение с подставленными аргументами в io.Writer.
// Поведение обработки недостающих/лишних аргументов аналогично Replace.
func (p *PlaceholderReplacer) ReplaceTo(wr io.Writer, args []any) error {
	if p.message == "" {
		return nil
	}

	if len(p.placeholders) == 0 {
		_, err := io.WriteString(wr, p.message)

		return err //nolint:wrapcheck
	}

	_, err := strings.NewReplacer(p.merge(args)...).WriteString(wr, p.message)

	return err //nolint:wrapcheck
}

// CountArgs - возвращает количество плейсхолдеров, ожидаемых сообщением.
func (p *PlaceholderReplacer) CountArgs() int {
	return len(p.placeholders)
}

func (p *PlaceholderReplacer) merge(args []any) (keyValues []string) {
	keyValues = make([]string, 0, len(p.placeholders)*2)

	for i, placeholder := range p.placeholders {
		value := missingArg

		if i < len(args) {
			value = conv.String(args[i])
		}

		keyValues = append(keyValues, placeholder, value)
	}

	return keyValues
}
