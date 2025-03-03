package test

import (
	"io"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

const (
	someUsagePattern      = "usage"
	usageCommandPattern   = `(go run ("[^"]+"|'[^']+|[^"' ]+))`
	inputNotFoundPattern  = `input not found`
	lookupNotFoundPattern = `lookup not found`
)

var helpFlags = [...]string{
	"-h",
	//	"-help",
	//	"--help",
}

var usageKeywordPatterns = [...]string{
	"input",
	"output",
	"airport|lookup",
}

// hasSomeUsage checks if the console output has "some" usage information.
// It doesn't have to be exhaustive or valid, but there is at least something.
func hasSomeUsage(output string) bool {
	return regexp.MustCompile(someUsagePattern).MatchString(output)
}

// TestNoArgs validates that the program prints some usage message to stdout
// when no flags and arguments were provided, and exits with no error.
func TestNoArgs(t *testing.T) {
	if output := run(t); !hasSomeUsage(output) {
		t.Fatalf("Output has no usage:\n%s", output)
	}
}

// TestHelpFlag validates that the program prints some usage message to stdout
// when help flag is provided as first argument. Global variable `helpFlags` holds
// the list of flags that must work as help flag.
func TestHelpFlag(t *testing.T) {
	for _, flag := range helpFlags {
		if output, err := runUnhandled(t, flag); err != nil {
			t.Fatalf("Program exited with error when help flag '%s' was provided: %s", flag, err)
		} else if !hasSomeUsage(output) {
			t.Fatalf("Output has no usage when flag '%s' is provided:\n%s", flag, output)
		}
	}
}

// TestInvalidFlag validates that the program exits with an error code when an unknown
// flag is provided as first argument. The program is expected to exit with an error.
func TestInvalidFlag(t *testing.T) {
	if _, err := runUnhandled(t, "-"+getUniqueName()); err == nil {
		t.Fatalf("Error expected")
	}
}

// TestUsage validates whole usage message printed to stdout with '-h' as first argument.
// Global list of regexp patters `usageKeywordPatterns` is used for pattern matching.
// Usage messaged is expected to match all of the patterns.
func TestUsage(t *testing.T) {
	output := run(t, "-h")

	// Iterate over all patterns
	for _, pattern := range usageKeywordPatterns {
		// It's your own responsibility to make sure the patterns compile.
		if !regexp.MustCompile(pattern).MatchString(output) {
			t.Errorf(`Cannot match "%s" in output`, pattern)

			// Do not call fatal right away to provide more info in single run
			// Let the tester see all unmatched patterns
			t.Fail()
		}
	}

	if t.Failed() {
		t.Error("Output:")
		t.Error(output)
		t.Fail()
	}
}

// TestInvalidFlag validates that the program exits with an error code when not enough
// arguments were passed. The program is expected to exit with an error.
func TestTooFewArgs(t *testing.T) {
	// Get two temporary files for testing purposes
	if err := withTempFile2(t.TempDir(), func(file1, file2 *os.File) {
		cases := [][]string{
			// Only one file is provided
			{file1.Name()},
			// Only two files are provided
			{file1.Name(), file2.Name()},
		}

		// Run all cases
		for _, c := range cases {
			// Validate that there is some error
			if o, err := runUnhandled(t, c...); err == nil {
				t.Fatalf("Got no error with %d args\nOutput:\n%s", len(c), o)
			}
		}
	}); err != nil {
		t.Fatal("Unexpected error: ", err)
	}
}

// TestRunCommandFromUsage tries to get a command from usage message and execute it.
// The pattern for extracting the command from the usage message is stored in constant
// `usageCommandPattern`, it is supposed to have at least one capturing group for the
// actual command (due to go's regexp lib not supporting lookbehinds), the command is
// extracted from the first capturing group. The extracted string is split by space to
// be passed as arguments to exec.Command.
//
// The pattern is expected to extract only the program running command, no arguments
// should be a part of it. Three temporary files are provided to the program as arguments.
// Input and output files are empty, lookup file is filled with arbitrary valid data.
//
// The program is expected to run without an error. Any content written to the output
// file is not validated. The purpose of this test is to validate that the command
// provided in the usage message can be ran successfully.
//
// If the pattern is empty, whole usage message is expected to be valid command.
//
// Example:
//   Usage message:    'Usage: go run myprogram.go [opts] arg'
//   Matching pattern: 'Usage: (go run [\w.]+)'
//   Matched string:   'go run myprogram.go'
//   Executed command: 'go run myprogram.go tempfile1 tempfile2 tempfile3'
func TestRunCommandFromUsage(t *testing.T) {
	// Get usage message
	usage := run(t, "-h")

	var commandString string

	// If pattern is empty, whole usage is valid command
	if usageCommandPattern == "" {
		commandString = usage
	} else {
		match := regexp.MustCompile(usageCommandPattern).FindStringSubmatch(usage)
		if len(match) >= 2 {
			commandString = match[1]
		}
	}

	// Remove extra space around
	commandString = strings.TrimSpace(commandString)

	// If the command is empty
	if commandString == "" {
		t.Fatalf("Command was not found in usage '%s'", usage)
	}

	// Extract executing program name and arguments
	// e.g. for 'go run .' the actual program is 'go'
	parts := strings.Split(commandString, " ")
	commandName, args := parts[0], parts[1:]

	// Run the program with three temporary files used as arguments
	if err := withTempFile3(t.TempDir(), func(inputFile, outputFile, lookupFile *os.File) {
		args = append(args, inputFile.Name(), outputFile.Name(), lookupFile.Name())
		command := exec.Command(commandName, args...)

		// Write basic lookup data to the lookup file
		if _, err := lookupFile.WriteString(basicLookup); err != nil {
			t.Fatal(err)
		}

		// Execute the command and check if it exits with an error
		if err := command.Run(); err != nil {
			t.Fatalf(
				"Could not run command from usage: '%s'\n%s",
				strings.Join(append([]string{commandName}, args...), " "),
				err,
			)
		}
	}); err != nil {
		t.Fatal("Unexpected error: ", err)
	}
}

// TestFileNotExist validates that the program exits with an error
// when any of the input files (input, lookup) do not exist or are
// not accessible by the program.
func TestFileNotExist(t *testing.T) {
	// Make just one temporary file to use as existing file for one
	// of the arguments
	if err := withTempFile(t.TempDir(), func(file *os.File) {
		t.Run("Input", func(t *testing.T) {
			// Run command and make sure it exits with an error
			if output, err := runUnhandled(t, path.Join(t.TempDir(), getUniqueName()), file.Name(), file.Name()); err == nil {
				t.Error("Expected error")
				t.Fail()
			} else if !regexp.MustCompile(inputNotFoundPattern).MatchString(output) {
				t.Errorf("'%s' pattern not found in output:\n%s", inputNotFoundPattern, output)
				t.Fail()
			}
		})

		t.Run("Lookup", func(t *testing.T) {
			// Run command and make sure it exits with an error
			if output, err := runUnhandled(t, file.Name(), file.Name(), path.Join(t.TempDir(), getUniqueName())); err == nil {
				t.Error("Expected error")
				t.Fail()
			} else if !regexp.MustCompile(lookupNotFoundPattern).MatchString(output) {
				t.Errorf("'%s' pattern not found in output:\n%s", lookupNotFoundPattern, output)
				t.Fail()
			}
		})
	}); err != nil {
		t.Fatal("Unexpected error: ", err)
	}
}

func TestOutputFileGetsCreated(t *testing.T) {
	//	sdad
}

// TestOutputFileCreatesDirs validates that the program is able to create all
// necessary directories preceeding the output file if they do not exist yet.
// The test expects the program to create 4 subdirectories A, B, C and D in
// already existing parent directory: /path/to/parent/A/B/C/D/output.txt.
//
// The test writes simple test string into input file and tries to read it
// unchanged from the output file. The output is trimmed in case of extra space
// produced by the program.
func TestOutputFileCreatesDirs(t *testing.T) {
	if err := withTempFile2(t.TempDir(), func(inputFile, lookupFile *os.File) {
		expected := "Test"

		writeAndCloseFile(t, inputFile, expected)
		writeAndCloseFile(t, lookupFile, basicLookup)

		// Generate random path to the ouput file
		outputFilePath := path.Join(t.TempDir(), getUniqueName(), getUniqueName(), getUniqueName(), getUniqueName()+".txt")

		// Run the program and check if it exits with an error
		if _, err := runUnhandled(t, inputFile.Name(), outputFilePath, lookupFile.Name()); err != nil {
			// Try to access output file
			f, e := os.Open(outputFilePath)
			if e == nil {
				// The output file was created successfully but something went wrong
				// Perhaps, there was an error with the input files.
				defer f.Close()
				t.Fatal("Program exited with an error, but the output file was created anyway")
			}

			// Output file failed to be created
			t.Fatal("Failed to run command: ", err)
		}

		// Try to access output file
		ofile, err := os.Open(outputFilePath)
		if err != nil {
			// Program exited successfully but the output file was not created
			// Where did it write the output???
			t.Fatal("Program exited without an error, but the output file was not created")
		}
		defer ofile.Close()

		// Read the output file
		actual, err := io.ReadAll(ofile)
		if err != nil {
			t.Fatal("Failed to read from output file")
		}

		// Compare the actual output with expected output
		if strings.TrimSpace(string(actual)) != expected {
			t.Fatalf("File was created successfully, but the output does not match: '%s'", string(actual))
		}
	}); err != nil {
		t.Fatal("Unexpected error: ", err)
	}
}

// caseFile holds info about a file of the case and expectations of a file
// from running the case.
type caseFile struct {
	description  string
	path         string
	removeBefore bool
	removeAfter  bool
	failOnExist  bool
	handle       *os.File
}

// runTestCaseOutputAffected validates files according to rules in caseFile
// when running the program with these files as arguments.
func runTestCaseOutputAffected(t *testing.T, files ...caseFile) {
	// Remove all files that had to be removed prematurely
	for _, file := range files {
		if file.removeBefore {
			var removed bool
			if file.handle != nil {
				removed = removeFile(file.handle)
			} else {
				removed = removeFileByPath(file.path)
			}
			if !removed {
				t.Fail()
			}
		}
	}

	if t.Failed() {
		t.Fatalf("Essential file failed to be removed")
	}

	// Populate args
	var args []string
	for _, file := range files {
		args = append(args, file.path)
	}

	// Run the program
	_, _ = runUnhandled(t, args...)
	//	output, _err := runUnhandled(t, args...)
	//	fmt.Println(output)
	//	fmt.Println(_err)

	// Remove all files that had to be removed after execution
	for _, file := range files {
		// Check if any files exist that should not exist
		if file.failOnExist {
			if fileExists(file.path) {
				// File exists, but should not
				t.Errorf("File exists when it sould not (%s)", file.description)
				t.Fail()
			}
		}

		// Remove all files that need to be removed
		if file.removeAfter {
			if file.handle != nil {
				removeFile(file.handle)
			} else {
				removeFileByPath(file.path)
			}
		}
	}
}

// TestOutputIsNotCreatedOnError validates that the output file is not created in case
// the program encounters any errors during execution.
//
// The program is expected to not create nor alter the content of the output file
// if any error occurs during execution. In addition to this, the program must not
// produce any ouput in case -h is passed.
func TestOutputIsNotCreatedOnError(t *testing.T) {
	// Test if output file is created whenever -h flag is passed
	t.Run("HelpArg", func(t *testing.T) {
		if err := withMockFiles(t.TempDir(), "test", basicLookup, func(inputFile, outputFile, lookupFile *os.File) {
			helpArg := caseFile{"helpFlag", "-h", false, false, false, nil}

			files := []caseFile{
				{"input", inputFile.Name(), false, false, false, inputFile},
				{"output", outputFile.Name(), true, true, true, outputFile},
				{"lookup", lookupFile.Name(), false, false, false, lookupFile},
			}

			for helpArgPos := 0; helpArgPos < 4; helpArgPos++ {
				var args []caseFile
				args = append(args, files[:helpArgPos]...)
				args = append(args, helpArg)
				args = append(args, files[helpArgPos:]...)
				runTestCaseOutputAffected(t, args...)
			}
		}); err != nil {
			t.Fatal("Unexpected error: ", err)
		}
	})

	// Test if output file is created whenever only two files
	// are provided to program instead of three
	t.Run("NoLookupArg", func(t *testing.T) {
		if err := withMockFiles(t.TempDir(), "test", "", func(inputFile, outputFile, lookupFile *os.File) {
			runTestCaseOutputAffected(t, []caseFile{
				{"input", inputFile.Name(), false, false, false, inputFile},
				{"output", outputFile.Name(), true, false, true, outputFile},
			}...)
		}); err != nil {
			t.Fatal("Unexpected error: ", err)
		}
	})

	// Test when input file does not exist
	t.Run("InputNotExist", func(t *testing.T) {
		if err := withMockFiles(t.TempDir(), "", basicLookup, func(inputFile, outputFile, lookupFile *os.File) {
			runTestCaseOutputAffected(t, []caseFile{
				{"input", inputFile.Name(), true, true, false, inputFile},
				{"output", outputFile.Name(), true, false, true, outputFile},
				{"lookup", lookupFile.Name(), false, false, false, lookupFile},
			}...)
		}); err != nil {
			t.Fatal("Unexpected error: ", err)
		}
	})

	// Test when lookup file does not exist
	t.Run("LookupNotExist", func(t *testing.T) {
		if err := withMockFiles(t.TempDir(), "test", "", func(inputFile, outputFile, lookupFile *os.File) {
			runTestCaseOutputAffected(t, []caseFile{
				{"input", inputFile.Name(), false, false, false, inputFile},
				{"output", outputFile.Name(), true, false, true, outputFile},
				{"lookup", lookupFile.Name(), true, true, false, lookupFile},
			}...)
		}); err != nil {
			t.Fatal("Unexpected error: ", err)
		}
	})

	// Test when lookup data is malformed
	t.Run("MalformedLookup", func(t *testing.T) {
		cases := []string{
			// Malformed header
			"name,iso_country,municipality,icao_code,iata_code" + lookupBasicBody,
			"name,iso_country,municipality,icao_code,coordinates" + lookupBasicBody,
			"name,iso_country,municipality,iata_code,coordinates" + lookupBasicBody,
			"name,iso_country,icao_code,iata_code,coordinates" + lookupBasicBody,
			"name,municipality,icao_code,iata_code,coordinates" + lookupBasicBody,
			"iso_country,municipality,icao_code,iata_code,coordinates" + lookupBasicBody,
			// Malformed content
			basicLookup + "\n,AU,Sydney,AAAA,BBB,\"160.05499267578, -9.4280004501343\"",
		}

		const input = "#BBB"

		for i, c := range cases {
			if !t.Run(strconv.FormatInt(int64(i+1), 10), func(t *testing.T) {
				if err := withMockFiles(t.TempDir(), input, c, func(inputFile, outputFile, lookupFile *os.File) {
					runTestCaseOutputAffected(t, []caseFile{
						{"input", inputFile.Name(), false, false, false, inputFile},
						{"output", outputFile.Name(), true, true, true, outputFile},
						{"lookup", lookupFile.Name(), false, false, false, outputFile},
					}...)
				}); err != nil {
					t.Fatal("Unexpected error: ", err)
				}
			}) {
				t.Error("Failed with the following lookup: ", c)
				t.Fail()
			}
		}
	})
}

// TestOutputIsNotAlteredOnError validates that the output file is not altered in case
// the program encounters any errors during execution.
//
// The program is expected to not create nor alter the content of the output file
// if any error occurs during execution.
func TestOutputIsNotAlteredOnError(t *testing.T) {

}

// TestInputIsNotAltered validates that the input file is not altered.
// The program is expected to never alter the input file.
func TestInputIsNotAltered(t *testing.T) {

}

// TestLookupIsNotAltered validates that the lookup file is not altered.
// The program is expected to never alter the input file.
func TestLookupIsNotAltered(t *testing.T) {

}
