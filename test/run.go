package test

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"unicode"
)

const cwd = "../"

const printStdoutWhenMocking = false
const printCommand = false

func init() {
	err := os.Chdir(cwd)
	if err != nil {
		panic(err)
	}
}

func runUnhandled(t *testing.T, args ...string) (string, error) {
	if printCommand {
		t.Logf("Running command: go run . %s", strings.Join(args, " "))
	}
	cmd := exec.Command("go", append([]string{"run", "."}, args...)...)

	output, err := cmd.CombinedOutput()

	if err != nil {
		if !errors.As(err, new(*exec.ExitError)) {
			t.Fatal(err)
		}
	}

	return string(output), err
}

func run(t *testing.T, args ...string) string {
	output, err := runUnhandled(t, args...)
	if err != nil {
		t.Errorf("Output:\n%s", output)
		t.Fatalf("Expected to exit with code 0, instead %s", err)
	}
	return output
}

func runWithMockFiles(t *testing.T, mockInput, mockLookup, expectedOutput string, pretrimContent bool, maxLinesToPrint int) (actualOutput string) {
	if err := withMockFiles(t.TempDir(), mockInput, mockLookup, func(inputFile, outputFile, lookupFile *os.File) {
		stdout := run(t, inputFile.Name(), outputFile.Name(), lookupFile.Name())

		if printStdoutWhenMocking {
			t.Log(stdout)
		}

		data, err := os.ReadFile(outputFile.Name())
		if err != nil {
			t.Fatalf("Could not read from actualOutput file: %s", err)
		}

		actualOutput = strings.TrimRightFunc(string(data), unicode.IsSpace)

		if !compareOutputs(t, expectedOutput, actualOutput, pretrimContent, maxLinesToPrint) {
			t.Fatalf("Actual output of the program significantly differs from expected output")
		}
	}); err != nil {
		t.Fatal(err)
	}

	return
}

func compareOutputs(t *testing.T, expectedOutput, actualOutput string, pretrim bool, maxLinesToPrint int) bool {
	if pretrim {
		expectedOutput = strings.TrimSpace(expectedOutput)
		actualOutput = strings.TrimSpace(actualOutput)
	}

	// Best case
	if expectedOutput == actualOutput {
		return true
	}

	// Perform line by line comparison
	expectedLines := strings.Split(expectedOutput, "\n")
	actualLines := strings.Split(actualOutput, "\n")

	if !compareLinesLengths(t, expectedLines, actualLines) {
		n := countSameLines(&expectedLines, &actualLines)

		t.Error("Expected output:")
		printLines(t, expectedLines, n, maxLinesToPrint)

		// Blank space
		t.Error()

		t.Error("Actual output:")
		printLines(t, actualLines, n, maxLinesToPrint)

		t.Fail()
		return false
	}

	compareLines(t, expectedLines, actualLines, pretrim, maxLinesToPrint)

	return !t.Failed()
}

func compareLinesLengths(t *testing.T, expected, actual []string) bool {
	l1, l2 := len(expected), len(actual)

	if l1 == l2 {
		return true
	}

	if l1 < l2 {
		t.Errorf("Actual output (%d lines) is larger than expected (%d lines)", l1, l2)
	} else {
		t.Errorf("Actual output (%d lines) is shorter than expected (%d lines)", l1, l2)
	}

	t.Fail()
	return false
}

// This function expects both slices to be of the same length
func compareLines(t *testing.T, lines1, lines2 []string, pretrim bool, maxPrint int) bool {
	if len(lines1) != len(lines2) {
		t.Fatal("Slices are not the same length")
	}

	for i := 0; i < len(lines1) && maxPrint > 0; i++ {
		if !compareStrings(lines1[i], lines2[i], pretrim) {
			t.Errorf("Line %d differs:\nExpected: '%s'\nActual:   '%s'", i,
				pretrimString(lines1[i], pretrim),
				pretrimString(lines2[i], pretrim))
			t.Fail()
			maxPrint--
		}
	}

	return !t.Failed()
}

func pretrimString(s string, flag bool) string {
	if flag {
		return strings.TrimSpace(s)
	}
	return s
}

func compareStrings(s1, s2 string, pretrim bool) bool {
	s1 = pretrimString(s1, pretrim)
	s2 = pretrimString(s2, pretrim)

	return s1 == s2
}

func printLines(t *testing.T, lines []string, skip int, max int) {
	if skip > 0 {
		t.Error("...")
	}
	for i, line := range lines[skip:] {
		if i >= max {
			t.Error("...")
			break
		}
		t.Errorf("Line %d: %s", skip+i, line)
	}
}

func countSameLines(a, b *[]string) (n int) {
	for ; n < len(*a) && n < len(*b); n++ {
		if !compareStrings((*a)[n], (*b)[n], true) {
			break
		}
	}
	return
}
