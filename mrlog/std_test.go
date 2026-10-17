package mrlog_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mondegor/go-core/mrlog"
	"github.com/mondegor/go-core/mrlog/level"
)

// fatalCaseEnv - имя переменной окружения, по которой тест, перезапущенный
// в дочернем процессе, выполняет вызов, завершающий процесс через os.Exit.
const fatalCaseEnv = "MRLOG_TEST_FATAL_CASE"

type (
	// stderrLogger - логгер дочернего процесса, пишущий записи в os.Stderr.
	stderrLogger struct {
		basicLogger
	}

	// stderrLevelLogger - stderrLogger с методом Log.
	stderrLevelLogger struct {
		stderrLogger
	}
)

func (l *stderrLogger) Error(_ context.Context, msg string, args ...any) {
	_, _ = fmt.Fprintln(os.Stderr, "ERROR", msg, args)
}

func (l *stderrLevelLogger) Log(_ context.Context, lvl level.Enum, msg string, args ...any) {
	_, _ = fmt.Fprintln(os.Stderr, lvl.String(), msg, args)
}

func TestFatal(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		run     func()
		want    string
		notWant string // если задано, не должно встречаться в выводе
	}

	tests := []testCase{
		{
			name: "FatalError via Log",
			run:  func() { mrlog.FatalError(&stderrLevelLogger{}, "msg", "k", 1) },
			want: "FATAL msg [k 1]",
		},
		{
			name: "FatalError via Error",
			run:  func() { mrlog.FatalError(&stderrLogger{}, "msg", "k", 1) },
			want: "ERROR Fatal error: msg [k 1]",
		},
		{
			name: "Fatal with logger",
			run:  func() { mrlog.Fatal(&stderrLevelLogger{}, "msg", "k", 1) },
			want: "FATAL msg [k 1]",
		},
		{
			name: "Fatal with std log",
			run:  func() { mrlog.Fatal("plain", "message") },
			want: "plainmessage",
		},
		{
			name:    "Fatal with logger without message",
			run:     func() { mrlog.Fatal(&stderrLevelLogger{}) },
			want:    "&{", // log.Fatal(v...) печатает логгер как указатель на структуру
			notWant: "FATAL",
		},
		{
			name:    "Fatal with logger and not string message",
			run:     func() { mrlog.Fatal(&stderrLevelLogger{}, 123) },
			want:    "123",
			notWant: "FATAL",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// в дочернем процессе выполняется только его вариант, остальные пропускаются
			if fatalCase := os.Getenv(fatalCaseEnv); fatalCase != "" {
				if fatalCase == tt.name {
					tt.run()
				}

				return
			}

			cmd := exec.Command(os.Args[0], "-test.run=^TestFatal$")

			cmd.Env = append(os.Environ(), fatalCaseEnv+"="+tt.name)

			out, err := cmd.CombinedOutput()

			var exitErr *exec.ExitError

			require.ErrorAs(t, err, &exitErr)
			assert.Equal(t, 1, exitErr.ExitCode())
			assert.Contains(t, string(out), tt.want)

			if tt.notWant != "" {
				assert.NotContains(t, string(out), tt.notWant)
			}
		})
	}
}
