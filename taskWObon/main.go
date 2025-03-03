package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

var LOOKUPTABLE map[string]string // global variables
var err error

func main() {
	if len(os.Args) != 4 || os.Args[1] == "-h" || os.Args[2] == "-h" || os.Args[3] == "-h" {
		println("itinerary usage:")
		println("go run . ./input.txt ./output.txt ./airport-lookup.csv")
		return
	}

	inputPath := os.Args[1]
	outputPath := os.Args[2]
	lookupPath := os.Args[3]

	LOOKUPTABLE, err = loadAirportLookup(lookupPath) // loading lookup
	if err != nil {
		fmt.Println("Error loading airport lookup:", err) // error
		return
	}

	err = processItinerary(inputPath, outputPath) // converting codes and times
	if err != nil {
		fmt.Println("Error processing itinerary:", err)
		return
	}

	println("Itinerary processed successfully.")
}

// loading lookup
func loadAirportLookup(lookupPath string) (map[string]string, error) {
	loo, err := os.Open(lookupPath) // open lookup
	if err != nil {
		return nil, fmt.Errorf("Lookup not found") // error
	}
	defer loo.Close()

	lookupTable := make(map[string]string) // making a map

	scanner := bufio.NewScanner(loo)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ",") // spliting by ","
		for i := 0; i < 5/*i <= len(parts)*/; i++ {
			if parts[i] == "" {
				return nil, fmt.Errorf("Malformed airport lookup data") // error
			}
		}
		if len(parts) < 5 /*|| len(parts) > 10*/{
			return nil, fmt.Errorf("Malformed airport lookup data5") // error
		}

		name := parts[0]
		municipality := parts[2]
		iata := parts[4]
		icao := parts[3]

		if iata != "" {
			lookupTable["*#"+iata] = municipality // printing city (municipality)
		}
		if icao != "" {
			lookupTable["*##"+icao] = municipality // printing city (municipality)
		}
		if iata != "" {
			lookupTable["#"+iata] = name // printing name of airport
		}
		if icao != "" {
			lookupTable["##"+icao] = name // printing name of airport
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err // error
	}

	return lookupTable, nil
}

// Working with files
func processItinerary(inputPath string, outputPath string) error {
	inputFile, err := os.Open(inputPath) // Opening input
	if err != nil {
		return fmt.Errorf("Input not found") // Error
	}
	defer inputFile.Close()

	outputFile, err := os.Create(outputPath) // Creating output
	if err != nil {
		return fmt.Errorf("Error creating output file") // Error
	}
	defer outputFile.Close()

	scanner := bufio.NewScanner(inputFile) // Reading input
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("Error reading input file") // Error
	}

	var outputTemp string

	for scanner.Scan() { // scanning
		line := scanner.Text()             // line
		processedLine := processLine(line) // convering
		outputTemp += processedLine + "\n" // writing info into a string
	}

	result := trimLines(outputTemp)  // trimming string
	fmt.Fprintln(outputFile, result) //Writing string to file

	return nil
}

// Trim new lines and change \v \r \f to \n
func trimLines(text string) (result string) {
	text = strings.ReplaceAll(text, "\v", "\n")
	text = strings.ReplaceAll(text, "\f", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	regex := regexp.MustCompile(`\n\n+`) // compiling blank lines to 1 blank line
	result = regex.ReplaceAllString(text, "\n\n")
	return
}

// Converting #, ##, * and dates with times
func processLine(line string) string {
	// *# with 3 characters, *## with 4 characters and so on
	re := regexp.MustCompile(`\*\#([A-Z]{3})|\*\##([A-Z]{4})|#([A-Z]{3})|##([A-Z]{4})|D\(([^)]+)\)|T12\(([^)]+)\)|T24\(([^)]+)\)`)

	return re.ReplaceAllStringFunc(line, func(match string) string {
		switch {
		case strings.HasPrefix(match, "#") || strings.HasPrefix(match, "##"): // if # or ##
			if name, exists := LOOKUPTABLE[match]; exists { // returning name of airport
				return name
			}

		case strings.HasPrefix(match, "*#") || strings.HasPrefix(match, "*##"): // if *# or *##
			if municipality, exists := LOOKUPTABLE[match]; exists { // returning city (municipality)
				return municipality
			}

		case strings.HasPrefix(match, "D("): // if D(Date)
			return formatISODate(match) // converting Date from D(YYYY-MM-DDTHH:mmZ) to human readable

		case strings.HasPrefix(match, "T12("): // if T12(Time)
			return formatISOTime(match[4:len(match)-1], true) // converting Time from T12(YYYY-MM-DDTHH:mmZ) to human readable

		case strings.HasPrefix(match, "T24("): // if T24
			return formatISOTime(match[4:len(match)-1], false) // converting Time from T24(YYYY-MM-DDTHH:mmZ) to human readable
		}

		return match
	})
}

// Formatting Date
func formatISODate(isoDate string) string {
	parsedTime, err := time.Parse("D(2006-01-02T15:04Z)", isoDate) // if "Z" formatting
	if err == nil {
		return parsedTime.Format("02 Jan 2006")
	}

	parsedTimeWithOffset, err := time.Parse("D(2006-01-02T15:04-07:00)", isoDate) // if "02:00" formatting
	if err != nil {
		return isoDate
	}

	return parsedTimeWithOffset.Format("02 Jan 2006")
}

// Formatting Time
func formatISOTime(isoTime string, is12HourFormat bool) string {
	var t time.Time
	var formattedTime string
	var offset string
	parts := strings.Split(isoTime, "T") // splitting by "T"
	if len(parts) != 2 {
		return isoTime
	}

	if strings.HasSuffix(parts[1], "Z") { // if "Z"
		const customFormat = "2006-01-02T15:04Z"   // format
		t, err = time.Parse(customFormat, isoTime) // formatting
	} else {
		const customFormat = "2006-01-02T15:04-07:00" // if "02:00" format
		t, err = time.Parse(customFormat, isoTime)    // formatting
	}

	if err != nil {
		return isoTime
	}

	if strings.HasSuffix(parts[1], "Z") { // from "Z" fromatting to "00:00"
		offset = "(+00:00)"
	} else {
		offset = t.Format("(-07:00)") // else formatting
	}

	if is12HourFormat {
		formattedTime = t.Format("03:04PM") // if 12H format
	} else {
		formattedTime = t.Format("15:04") // if 24H format
	}

	return fmt.Sprintf("%s %s", formattedTime, offset) // printing human readable
}
