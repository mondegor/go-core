package mrmsg

import (
	"fmt"
	"io"
	"strings"
)

type (
	// FmtReplacer - подставляет аргументы в сообщение формата fmt (например: "%[1]s") через fmt.Sprintf.
	// Плейсхолдеры задают только количество ожидаемых аргументов.
	FmtReplacer struct {
		message      string
		placeholders []string
	}
)

// NewFmtReplacer - создаёт FmtReplacer для замены плейсхолдеров в сообщении.
// Параметры:
//   - message - строка формата для fmt.Sprintf (например: "Hello %[1]s, you are %[2]d");
//   - placeholders - исходные плейсхолдеры сообщения, их количество задаёт число аргументов.
func NewFmtReplacer(message string, placeholders []string) *FmtReplacer {
	return &FmtReplacer{
		message:      message,
		placeholders: placeholders,
	}
}

// Replace - подставляет аргументы в подготовленное сообщение через fmt.Sprintf.
// Сообщение всегда обрабатывается как строка формата fmt, даже без плейсхолдеров
// (например: "100%% done" → "100% done").
// Недостающие аргументы заменяются на "!MISSINGARG", лишние игнорируются.
func (p *FmtReplacer) Replace(args []any) (replacedMessage string, err error) {
	var buf strings.Builder

	if err = p.ReplaceTo(&buf, args); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ReplaceTo - записывает сообщение с подставленными аргументами в io.Writer.
// Поведение обработки сообщения и недостающих/лишних аргументов аналогично Replace.
func (p *FmtReplacer) ReplaceTo(wr io.Writer, args []any) error {
	if p.message == "" {
		return nil
	}

	_, err := fmt.Fprintf(wr, p.message, p.correctArgs(args)...)

	return err //nolint:wrapcheck
}

// CountArgs - возвращает количество плейсхолдеров, ожидаемых сообщением.
func (p *FmtReplacer) CountArgs() int {
	return len(p.placeholders)
}

func (p *FmtReplacer) correctArgs(args []any) []any {
	corrected := make([]any, len(p.placeholders))

	for i := copy(corrected, args); i < len(corrected); i++ {
		corrected[i] = missingArg
	}

	return corrected
}
