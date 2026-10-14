package mrmsg

import (
	"strings"
)

type (
	// MessageFormatter - преобразует плейсхолдеры в сообщении в формат,
	// подходящий для конкретного реплейсера (fmt.Sprintf, strings.Replacer и т.д.).
	// Извлекает плейсхолдеры и преобразует каждый через пользовательскую функцию formatter.
	MessageFormatter struct {
		extractor *PlaceholderExtractor
		formatter func(placeholder string, index int) (newPlaceholder string)
	}
)

// NewMessageFormatter - создаёт MessageFormatter для преобразования плейсхолдеров.
// Параметры:
//   - leftDelim, rightDelim - ограничители плейсхолдеров (например: "{{" и "}}");
//   - formatter - функция преобразования плейсхолдера в новый формат, принимает исходный плейсхолдер и его индекс;
//     если formatter не задан, плейсхолдеры не изменяются.
func NewMessageFormatter(leftDelim, rightDelim string, formatter func(placeholder string, index int) (newPlaceholder string)) *MessageFormatter {
	if formatter == nil {
		formatter = func(placeholder string, _ int) (newPlaceholder string) {
			return placeholder
		}
	}

	return &MessageFormatter{
		extractor: NewPlaceholderExtractor(leftDelim, rightDelim),
		formatter: formatter,
	}
}

// Format - преобразует плейсхолдеры в сообщении в формат для целевого реплейсера.
// Возвращает преобразованное сообщение и список новых плейсхолдеров.
// Например: "{name}" → "%[1]s" для fmt.Sprintf.
func (p *MessageFormatter) Format(message string) (formattedMessage string, newPlaceholders []string) {
	placeholders := p.extractor.Extract(message)

	if len(placeholders) == 0 {
		return message, placeholders
	}

	oldNew := make([]string, 0, len(placeholders)*2)

	for i := range placeholders {
		oldNew = append(oldNew, placeholders[i], p.formatter(placeholders[i], i))
		placeholders[i] = oldNew[len(oldNew)-1]
	}

	return strings.NewReplacer(oldNew...).Replace(message), placeholders
}
