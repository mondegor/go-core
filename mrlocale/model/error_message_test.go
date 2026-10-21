package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mondegor/go-core/mrlocale/model"
)

func TestErrorMessage(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		message     model.ErrorMessage
		translation string
	}

	tests := []testCase{
		{
			name:        "reason only",
			message:     model.ErrorMessage{Reason: "reason"},
			translation: "reason",
		},
		{
			name:        "reason and details",
			message:     model.ErrorMessage{Reason: "reason", Details: "details"},
			translation: "reason\n\ndetails",
		},
		{
			name:        "all parts",
			message:     model.ErrorMessage{Reason: "reason", Details: "details", ProblemURL: "https://host/problem"},
			translation: "reason\n\ndetails\n\nhttps://host/problem",
		},
		{
			name:        "problem url without details",
			message:     model.ErrorMessage{Reason: "reason", ProblemURL: "https://host/problem"},
			translation: "reason\n\n\n\nhttps://host/problem",
		},
		{
			name:        "empty",
			message:     model.ErrorMessage{},
			translation: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.translation, tt.message.String())
			assert.Equal(t, tt.message, model.ParseErrorMessage(tt.translation))
		})
	}
}

func TestParseErrorMessage_SeparatorInURL(t *testing.T) {
	t.Parallel()

	// всё после второго разделителя относится к ProblemURL
	got := model.ParseErrorMessage("reason\n\ndetails\n\nurl\n\ntail")
	assert.Equal(t, model.ErrorMessage{Reason: "reason", Details: "details", ProblemURL: "url\n\ntail"}, got)
}
