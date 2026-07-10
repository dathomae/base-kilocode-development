// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestSynthesizeConversions_Empty verifies that an empty slice returns an
// empty string.
func TestSynthesizeConversions_Empty(t *testing.T) {
	got := SynthesizeConversions(nil, "unitConversions", "convertUnit", "unit")
	if got != "" {
		t.Errorf("SynthesizeConversions(nil) = %q, want empty string", got)
	}
	got = SynthesizeConversions([]CILConversionPair{}, "unitConversions", "convertUnit", "unit")
	if got != "" {
		t.Errorf("SynthesizeConversions([]) = %q, want empty string", got)
	}
}

// TestSynthesizeConversions_UnitSinglePair verifies the structure of the
// emitted map and helper function for a single unit conversion pair.
func TestSynthesizeConversions_UnitSinglePair(t *testing.T) {
	pairs := []CILConversionPair{
		{SourceName: "Meter", TargetName: "Kilometer", SourceVar: "m", TargetVar: "km", Calculation: "(m / 1000)"},
	}
	code := SynthesizeConversions(pairs, "unitConversions", "convertUnit", "unit")

	if !strings.Contains(code, "var unitConversions") {
		t.Errorf("missing unitConversions variable")
	}
	if !strings.Contains(code, "func convertUnit(sourceName, targetName string, value float64) float64") {
		t.Errorf("missing convertUnit function")
	}
	if !strings.Contains(code, `"Meter"`) {
		t.Errorf("missing Meter source unit")
	}
	if !strings.Contains(code, `"Kilometer"`) {
		t.Errorf("missing Kilometer target unit")
	}
	if !strings.Contains(code, "func(m float64) float64 { return (m / 1000) }") {
		t.Errorf("missing anonymous conversion function with correct parameter and body")
	}
}

// TestSynthesizeConversions_UnitForwardAndInverse verifies that both forward
// and inverse pairs are emitted.
func TestSynthesizeConversions_UnitForwardAndInverse(t *testing.T) {
	pairs := []CILConversionPair{
		{SourceName: "Meter", TargetName: "Kilometer", SourceVar: "m", TargetVar: "km", Calculation: "(m / 1000)"},
		{SourceName: "Kilometer", TargetName: "Meter", SourceVar: "km", TargetVar: "m", Calculation: "(km * 1000)"},
	}
	code := SynthesizeConversions(pairs, "unitConversions", "convertUnit", "unit")

	if strings.Count(code, `"Meter"`) < 2 {
		t.Errorf("expected Meter to appear at least twice (source + target)")
	}
	if strings.Count(code, `"Kilometer"`) < 2 {
		t.Errorf("expected Kilometer to appear at least twice (source + target)")
	}
}

// TestSynthesizeConversions_UnitCompilesAndRuns verifies that the emitted Go
// code can be compiled and executes correctly.
func TestSynthesizeConversions_UnitCompilesAndRuns(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("'go' command not available in PATH")
	}

	pairs := []CILConversionPair{
		{SourceName: "Meter", TargetName: "Kilometer", SourceVar: "m", TargetVar: "km", Calculation: "(m / 1000)"},
		{SourceName: "Kilometer", TargetName: "Meter", SourceVar: "km", TargetVar: "m", Calculation: "(km * 1000)"},
	}
	conversionCode := SynthesizeConversions(pairs, "unitConversions", "convertUnit", "unit")

	// Wrap the conversion code in a tiny Go program.
	program := fmt.Sprintf(`package main

import "fmt"

%s

func main() {
	fmt.Printf("%%.0f\n", convertUnit("Meter", "Kilometer", 5000))
	fmt.Printf("%%.0f\n", convertUnit("Kilometer", "Meter", 3))
}
`, conversionCode)

	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(srcPath, []byte(program), 0644); err != nil {
		t.Fatalf("failed to write test program: %v", err)
	}

	cmd := exec.Command("go", "run", srcPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run generated program: %v\nOutput:\n%s", err, string(out))
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 output lines, got %d: %q", len(lines), string(out))
	}
	if lines[0] != "5" {
		t.Errorf("Meter->Kilometer(5000) = %s, want 5", lines[0])
	}
	if lines[1] != "3000" {
		t.Errorf("Kilometer->Meter(3) = %s, want 3000", lines[1])
	}
}

// TestSynthesizeConversions_MeasureEmpty verifies that an empty slice returns
// an empty string.
func TestSynthesizeConversions_MeasureEmpty(t *testing.T) {
	got := SynthesizeConversions(nil, "measureConversions", "convertMeasure", "measure")
	if got != "" {
		t.Errorf("SynthesizeConversions(nil) = %q, want empty string", got)
	}
	got = SynthesizeConversions([]CILConversionPair{}, "measureConversions", "convertMeasure", "measure")
	if got != "" {
		t.Errorf("SynthesizeConversions([]) = %q, want empty string", got)
	}
}

// TestSynthesizeConversions_MeasureSinglePair verifies the structure of the
// emitted map and helper function for a single measure conversion pair.
func TestSynthesizeConversions_MeasureSinglePair(t *testing.T) {
	pairs := []CILConversionPair{
		{SourceName: "Celsius", TargetName: "Fahrenheit", SourceVar: "c", TargetVar: "f", Calculation: "((c * 1.8) + 32)"},
	}
	code := SynthesizeConversions(pairs, "measureConversions", "convertMeasure", "measure")

	if !strings.Contains(code, "var measureConversions") {
		t.Errorf("missing measureConversions variable")
	}
	if !strings.Contains(code, "func convertMeasure(sourceName, targetName string, value float64) float64") {
		t.Errorf("missing convertMeasure function")
	}
	if !strings.Contains(code, `"Celsius"`) {
		t.Errorf("missing Celsius source measure")
	}
	if !strings.Contains(code, `"Fahrenheit"`) {
		t.Errorf("missing Fahrenheit target measure")
	}
	if !strings.Contains(code, "func(c float64) float64 { return ((c * 1.8) + 32) }") {
		t.Errorf("missing anonymous conversion function with correct parameter and body")
	}
}

// TestSynthesizeConversions_MeasureCompilesAndRuns verifies that the emitted
// Go code can be compiled and executes correctly.
func TestSynthesizeConversions_MeasureCompilesAndRuns(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("'go' command not available in PATH")
	}

	pairs := []CILConversionPair{
		{SourceName: "Celsius", TargetName: "Fahrenheit", SourceVar: "c", TargetVar: "f", Calculation: "((c * 1.8) + 32)"},
		{SourceName: "Fahrenheit", TargetName: "Celsius", SourceVar: "f", TargetVar: "c", Calculation: "((f - 32) / 1.8)"},
	}
	conversionCode := SynthesizeConversions(pairs, "measureConversions", "convertMeasure", "measure")

	program := fmt.Sprintf(`package main

import "fmt"

%s

func main() {
	fmt.Printf("%%.0f\n", convertMeasure("Celsius", "Fahrenheit", 0))
	fmt.Printf("%%.0f\n", convertMeasure("Fahrenheit", "Celsius", 212))
}
`, conversionCode)

	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(srcPath, []byte(program), 0644); err != nil {
		t.Fatalf("failed to write test program: %v", err)
	}

	cmd := exec.Command("go", "run", srcPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run generated program: %v\nOutput:\n%s", err, string(out))
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 output lines, got %d: %q", len(lines), string(out))
	}
	if lines[0] != "32" {
		t.Errorf("Celsius->Fahrenheit(0) = %s, want 32", lines[0])
	}
	if lines[1] != "100" {
		t.Errorf("Fahrenheit->Celsius(212) = %s, want 100", lines[1])
	}
}
