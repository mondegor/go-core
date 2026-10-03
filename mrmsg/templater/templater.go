package templater

import (
	"fmt"
	htmltemplate "html/template"
	"io"
	"strings"
	texttemplate "text/template"
)

const (
	leftDelimDefault  = "{{"
	rightDelimDefault = "}}"

	// при отсутствии переменной в data выводится пустая строка (а не "<no value>").
	missingKeyOption = "missingkey=zero"
)

// Режимы рендеринга шаблона.
const (
	ModeText Mode = iota // text/template: значения выводятся без изменений (по умолчанию)
	ModeHTML             // html/template: значения экранируются как HTML
)

type (
	// Templater - шаблонизатор сообщений на основе text/template или html/template
	// в зависимости от режима рендеринга.
	// Позволяет использовать синтаксис Go-шаблонов для подстановки параметров.
	Templater struct {
		leftDelim  string
		rightDelim string
		mode       Mode
	}

	// Mode - режим рендеринга шаблона.
	// Неизвестные значения обрабатываются как ModeHTML (с экранированием).
	Mode uint8

	executor interface {
		Execute(wr io.Writer, data any) error
	}
)

// NewTemplater - создаёт Templater с указанными ограничителями.
// Если leftDelim или rightDelim пусты, используются значения по умолчанию: "{{" и "}}".
func NewTemplater(leftDelim, rightDelim string, opts ...Option) *Templater {
	if leftDelim == "" {
		leftDelim = leftDelimDefault
	}

	if rightDelim == "" {
		rightDelim = rightDelimDefault
	}

	o := options{
		templater: &Templater{
			leftDelim:  leftDelim,
			rightDelim: rightDelim,
			mode:       ModeText,
		},
	}

	for _, opt := range opts {
		opt(&o)
	}

	return o.templater
}

// Render - формирует сообщение из шаблона, подставляя параметры из data.
// Параметры:
//   - message - шаблон с синтаксисом Go-шаблонов;
//   - data - карта имён параметров и их строковых значений;
func (p *Templater) Render(message string, data map[string]string) (string, error) {
	var buf strings.Builder

	if err := p.RenderTo(&buf, message, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// RenderTo - записывает сформированное сообщение из шаблона в io.Writer.
// Параметры:
//   - message - шаблон с синтаксисом Go-шаблонов;
//   - data - карта имён параметров и их строковых значений;
//
// Если сообщение не содержит ограничителей, записывается как есть без парсинга.
// Отсутствующие в data параметры выводятся пустой строкой.
func (p *Templater) RenderTo(wr io.Writer, message string, data map[string]string) error {
	if message == "" {
		return nil
	}

	if !strings.Contains(message, p.leftDelim) {
		_, err := wr.Write([]byte(message))

		return err //nolint:wrapcheck
	}

	t, err := p.parse(message)
	if err != nil {
		return fmt.Errorf("parse message '%s': %w", message, err)
	}

	if err = t.Execute(wr, data); err != nil {
		return fmt.Errorf("render message '%s': %w", message, err)
	}

	return nil
}

func (p *Templater) parse(message string) (executor, error) {
	switch p.mode {
	case ModeText:
		return texttemplate.New("").Delims(p.leftDelim, p.rightDelim).Option(missingKeyOption).Parse(message)
	default: // ModeHTML и неизвестные значения: безопаснее экранировать
		return htmltemplate.New("").Delims(p.leftDelim, p.rightDelim).Option(missingKeyOption).Parse(message)
	}
}
