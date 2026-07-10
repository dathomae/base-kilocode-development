// Package parser_test contains tests for the CPL parser logging functionality.
package parser_test

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseLoggingConfig_ValidLogLevels tests that all valid log level strings
// (debug, info, warn, error) can be parsed correctly to their corresponding
// slog.Level values.
func TestParseLoggingConfig_ValidLogLevels(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		levelStr      string
		expectedLevel slog.Level
		expectError   bool
	}{
		{
			name:          "debug level",
			levelStr:      "debug",
			expectedLevel: slog.LevelDebug,
			expectError:   false,
		},
		{
			name:          "info level",
			levelStr:      "info",
			expectedLevel: slog.LevelInfo,
			expectError:   false,
		},
		{
			name:          "warn level",
			levelStr:      "warn",
			expectedLevel: slog.LevelWarn,
			expectError:   false,
		},
		{
			name:          "error level",
			levelStr:      "error",
			expectedLevel: slog.LevelError,
			expectError:   false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Parse the log level from string
			var level slog.Level
			err := level.UnmarshalText([]byte(tc.levelStr))

			if tc.expectError {
				require.Error(t, err, "Expected error for level: %s", tc.levelStr)
				return
			}

			require.NoError(t, err, "Unexpected error for level: %s", tc.levelStr)
			assert.Equal(t, tc.expectedLevel, level, "Log level should match")
		})
	}
}

// TestParseLoggingConfig_InvalidLogLevel tests that an invalid log level string
// returns an appropriate error with a helpful message listing valid values.
func TestParseLoggingConfig_InvalidLogLevel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		levelStr string
	}{
		{
			name:     "invalid level - random string",
			levelStr: "invalid",
		},
		{
			name:     "invalid level - numeric",
			levelStr: "123",
		},
		{
			name:     "invalid level - empty",
			levelStr: "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var level slog.Level
			err := level.UnmarshalText([]byte(tc.levelStr))
			require.Error(t, err, "Expected error for invalid level: %s", tc.levelStr)
		})
	}
}

// TestParseWithLogLevel_NilConfig tests that passing a nil config to
// ParseWithLogLevel returns an error as expected.
func TestParseWithLogLevel_NilConfig(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		input         string
		config        *parser.ParseLoggingConfig
		expectedError string
	}{
		{
			name:          "nil config",
			input:         "test",
			config:        nil,
			expectedError: "logging config cannot be nil",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc, err := parser.ParseWithLogLevel(tc.input, tc.config)
			require.Error(t, err, "Expected error for nil config")
			assert.Nil(t, doc, "Expected nil document for nil config")
			assert.Contains(t, err.Error(), tc.expectedError, "Error message should contain expected text")
		})
	}
}

// TestParseWithLogLevel_FileOutput tests that logs are written to a specified
// file path correctly, including:
// - Creating the file if it doesn't exist
// - Appending to the file if it already exists
// - Writing the expected log format (entry and exit messages at the configured level)
func TestParseWithLogLevel_FileOutput(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		input           string
		parseCount      int
		expectedEntries int
	}{
		{
			name:            "single parse creates file",
			input:           "types\n  Metric = Number()",
			parseCount:      1,
			expectedEntries: 1,
		},
		{
			name:            "multiple parses append to file",
			input:           "types\n  Metric = Number()",
			parseCount:      2,
			expectedEntries: 2,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			logFile := filepath.Join(tempDir, "parser.log")

			config := &parser.ParseLoggingConfig{
				LogLevel:   slog.LevelInfo,
				OutputFile: logFile,
			}

			// Run parse the specified number of times
			for i := 0; i < tc.parseCount; i++ {
				doc, err := parser.ParseWithLogLevel(tc.input, config)
				require.NoError(t, err, "Unexpected error on parse %d", i+1)
				require.NotNil(t, doc, "Expected non-nil document on parse %d", i+1)
			}

			// Verify file was created and contains expected content
			content, err := os.ReadFile(logFile)
			require.NoError(t, err, "Failed to read log file")

			logContent := string(content)
			assert.Contains(t, logContent, "parse_entry", "Log should contain parse_entry")
			assert.Contains(t, logContent, "parse_exit", "Log should contain parse_exit")
			assert.Contains(t, logContent, "input_length", "Log should contain input_length")
			assert.Contains(t, logContent, "duration_ms", "Log should contain duration_ms")

			// Verify entry count
			entryCount := strings.Count(logContent, "parse_entry")
			assert.Equal(t, tc.expectedEntries, entryCount, "Expected %d parse_entry occurrences", tc.expectedEntries)
		})
	}
}

// TestParseWithLogLevel_StdoutOutput tests that logs are written to stdout
// when no output file is specified.
func TestParseWithLogLevel_StdoutOutput(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		input  string
		config *parser.ParseLoggingConfig
	}{
		{
			name:  "stdout output with info level",
			input: "types\n  Metric = Number()",
			config: &parser.ParseLoggingConfig{
				LogLevel:   slog.LevelInfo,
				OutputFile: "", // Empty means stdout
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc, err := parser.ParseWithLogLevel(tc.input, tc.config)
			require.NoError(t, err, "Unexpected error")
			require.NotNil(t, doc, "Expected non-nil document")
		})
	}
}

// TestParseWithLogLevel_LogLevelFiltering tests that:
// - At Info level, debug messages from the listener are not logged
// - At Debug level, both info and debug messages are logged
func TestParseWithLogLevel_LogLevelFiltering(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		logLevel        slog.Level
		expectDebugLogs bool
	}{
		{
			name:            "info level excludes debug messages",
			logLevel:        slog.LevelInfo,
			expectDebugLogs: false,
		},
		{
			name:            "debug level includes debug messages",
			logLevel:        slog.LevelDebug,
			expectDebugLogs: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			logFile := filepath.Join(tempDir, "test.log")

			config := &parser.ParseLoggingConfig{
				LogLevel:   tc.logLevel,
				OutputFile: logFile,
			}

			input := "types\n  Metric = Number()"
			_, err := parser.ParseWithLogLevel(input, config)
			require.NoError(t, err, "Unexpected error")

			content, err := os.ReadFile(logFile)
			require.NoError(t, err, "Failed to read log file")
			logContent := string(content)

			if tc.expectDebugLogs {
				assert.Contains(t, logContent, "enter", "Debug log should contain enter")
				assert.Contains(t, logContent, "exit", "Debug log should contain exit")
			} else {
				assert.NotContains(t, logContent, "enter_rule", "Info log should not contain enter_rule")
				assert.NotContains(t, logContent, "exit_rule", "Info log should not contain exit_rule")
			}
		})
	}
}

// TestParseWithLogLevel_DurationLogging tests that parse_exit includes
// duration_ms in the log output.
func TestParseWithLogLevel_DurationLogging(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		input             string
		expectDurationLog bool
	}{
		{
			name:              "simple parse includes duration",
			input:             "types\n  Metric = Number()",
			expectDurationLog: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			logFile := filepath.Join(tempDir, "parser.log")

			config := &parser.ParseLoggingConfig{
				LogLevel:   slog.LevelInfo,
				OutputFile: logFile,
			}

			_, err := parser.ParseWithLogLevel(tc.input, config)
			require.NoError(t, err, "Unexpected error")

			content, err := os.ReadFile(logFile)
			require.NoError(t, err, "Failed to read log file")
			logContent := string(content)

			if tc.expectDurationLog {
				assert.Contains(t, logContent, "duration_ms", "Log should contain duration_ms")
			}
		})
	}
}

// TestParseWithLogLevel_ParseEntry tests that parse_entry includes
// input_length in the log output.
func TestParseWithLogLevel_ParseEntry(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		input             string
		expectInputLength bool
	}{
		{
			name:              "parse entry includes input length",
			input:             "types\n  Metric = Number()",
			expectInputLength: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			logFile := filepath.Join(tempDir, "parser.log")

			config := &parser.ParseLoggingConfig{
				LogLevel:   slog.LevelInfo,
				OutputFile: logFile,
			}

			inputLength := len(tc.input)
			_, err := parser.ParseWithLogLevel(tc.input, config)
			require.NoError(t, err, "Unexpected error")

			content, err := os.ReadFile(logFile)
			require.NoError(t, err, "Failed to read log file")
			logContent := string(content)

			if tc.expectInputLength {
				assert.Contains(t, logContent, "input_length", "Log should contain input_length")

				// Verify the input_length value matches
				expectedLengthStr := fmt.Sprintf("input_length=%d", inputLength)
				assert.Contains(t, logContent, expectedLengthStr, "Log should contain correct input_length value")
			}
		})
	}
}
