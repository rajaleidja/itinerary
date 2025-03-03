package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

var LOOKUPTABLE map[string]string // global variables
var err error //error

var (
	helpFlag     bool
	bonusFlag    bool
	columnNumber int
)

func init() {
	flag.BoolVar(&helpFlag, "h", false, "Show help message")
	flag.BoolVar(&helpFlag, "help", false, "Show help message")

	flag.BoolVar(&bonusFlag, "b", false, "Enable bonus mode")
	flag.BoolVar(&bonusFlag, "bonus", false, "Enable bonus mode")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\033[32mUsage:\033[0m \033[34mgo run . \033[33m[-h help] [-b bonus]\033[0m \033[34m[INPUT FILE] [OUTPUT FILE] [LOOKUP FILE]\033[0m\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(os.Stderr, "  \033[35m-%s, --%s \033[0m- %s\n", f.Name, f.Name, f.Usage)
		})
		fmt.Println("  \033[33mEXAMPLE: go run . ./input.txt ./output.txt ./airport-lookup.csv\033[0m")
		fmt.Println("  \033[33mEXAMPLE: go run . \033[31m-b\033[33m ./input.txt ./output.txt ./airport-lookup.csv\033[0m")
	}
}

func processPositionalArguments(args []string) {
	fmt.Printf("\033[32mPositional arguments: %v\n \033[31mWANT: [INPUT] [OUTPUT] [LOOKUP]!\n\033[0m", args)
}

func main() {
	if len(os.Args) == 1 {
		println("itinerary usage:")
	 	println("go run . ./input.txt ./output.txt ./airport-lookup.csv")
		return
	}
	flag.Parse()

	if helpFlag && !bonusFlag{
		println("itinerary usage:")
	 	println("go run . ./input.txt ./output.txt ./airport-lookup.csv")
		return
	}
	if helpFlag && bonusFlag {
		flag.Usage()
		return
	}
	if bonusFlag {
		fmt.Printf("\033[32mBonus mode enabled\033[0m\n")
		processPositionalArguments(flag.Args())
	} else {
		fmt.Println("Normal mode, USE: \"\033[34mgo run . \033[33m-h -b\033[0m\" to see more")
	}

	 if len(os.Args) < 4 || len(os.Args) > 5 {
	 	println("itinerary usage:")
	 	println("go run . ./input.txt ./output.txt ./airport-lookup.csv")
	 	return
	 }

	inputPath := flag.Args()[0]
	outputPath := flag.Args()[1]
	lookupPath := flag.Args()[2]

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

	println("\033[32mItinerary processed successfully.\033[0m")

}

// loading lookup
func loadAirportLookup(lookupPath string) (map[string]string, error) {
	loo, err := os.Open(lookupPath) // open lookup
	if err != nil {
		return nil, fmt.Errorf("\033[31mLookup not found\033[0m") // error
	}
	defer loo.Close()

	lookupTable := make(map[string]string) // making a map
	chuits := 0
	chuits1 := 2
	chuits2 := 4
	chuits3 := 3
	if bonusFlag {
		
		print("enter number of column \"name\", DEFAULT = 0\n")
		fmt.Scanln(&chuits)
		print("enter number of column \"city\", DEFAULT = 2\n")
		fmt.Scanln(&chuits1)
		print("enter number of column \"IATA code\", DEFAULT = 4\n")
		fmt.Scanln(&chuits2)
		print("enter number of column \"ICAO code\", DEFAULT = 3\n")
		fmt.Scanln(&chuits3)
	}
	scanner := bufio.NewScanner(loo)
	
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ",") // spliting by ","
		for i := 0; i < 5; i++ {
			print(parts[i]," ", *&parts[i], "\n")
			if parts[i] == "" {
				return nil, fmt.Errorf("\033[31mMalformed airport lookup data\033[0m") // error
			}
		}
		if len(parts) <= 5 {
			return nil, fmt.Errorf("\033[31mMalformed airport lookup! data\033[0m") // error
		}

		name := parts[chuits] 			// 0 chuits
		municipality := parts[chuits1]  // 2 chuits1
		iata := parts[chuits2] 			// 4 chuits2
		icao := parts[chuits3] 			// 3 chuits3

		if iata != "" {
			lookupTable["*#"+iata] = "\033[34m"+municipality+ "\033[0m" // printing city (municipality)
		}
		if icao != "" {
			lookupTable["*##"+icao] = "\033[34m"+municipality+ "\033[0m" // printing city (municipality)
		}
		if iata != "" {
			lookupTable["#"+iata] = "\033[36m"+name+"\033[0m" // printing name of airport
		}
		if icao != "" {
			lookupTable["##"+icao] = "\033[36m"+name+"\033[0m" // printing name of airport
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
		return fmt.Errorf("\033[31mInput not found\033[0m") // Error
	}
	defer inputFile.Close()

	outputFile, err := os.Create(outputPath) // Creating output
	if err != nil {
		return fmt.Errorf("\033[31mError creating output file\033[0m") // Error
	}
	defer outputFile.Close()

	scanner := bufio.NewScanner(inputFile) // Reading input
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("\033[31mError reading input file\033[0m") // Error
	}

	var outputTemp string

	for scanner.Scan() { // scanning
		line := scanner.Text()             // line
		processedLine := processLine(line) // convering
		outputTemp += processedLine + "\n" // writing info into a string
	}
	result := trimLines(outputTemp)  // trimming string
	if bonusFlag{
		fmt.Println(result) //
	}
	result = trimColor(result)
	fmt.Fprintln(outputFile, result[:len(result)-1]) //Writing string to file

	return nil
}

func trimColor(text string) string {
	// text color
	text = strings.ReplaceAll(text, "\033[0m", "")
	text = strings.ReplaceAll(text, "\033[30m", "")
	text = strings.ReplaceAll(text, "\033[34m", "")
	text = strings.ReplaceAll(text, "\033[35m", "")
	text = strings.ReplaceAll(text, "\033[36m", "")
	text = strings.ReplaceAll(text, "\033[37m", "")
	text = strings.ReplaceAll(text, "\033[31m", "")
	text = strings.ReplaceAll(text, "\033[32m", "")
	text = strings.ReplaceAll(text, "\033[33m", "")
	// fon color
	text = strings.ReplaceAll(text, "\033[40m", "")
	text = strings.ReplaceAll(text, "\033[41m", "")
	text = strings.ReplaceAll(text, "\033[42m", "")
	text = strings.ReplaceAll(text, "\033[43m", "")
	text = strings.ReplaceAll(text, "\033[44m", "")
	text = strings.ReplaceAll(text, "\033[45m", "")
	text = strings.ReplaceAll(text, "\033[46m", "")
	text = strings.ReplaceAll(text, "\033[47m", "")
	// bold
	text = strings.ReplaceAll(text, "\033[1m", "")
	text = strings.ReplaceAll(text, "\033[22m", "")
	return text
}

// Trim new lines and change \v \r \f to \n
func trimLines(text string) (result string) {
	text = strings.ReplaceAll(text, "\v", "\n")
	text = strings.ReplaceAll(text, "\f", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.ReplaceAll(text, " \n", "\n")
	
	pattern := regexp.MustCompile(` \s+`)
    text = pattern.ReplaceAllString(text, " ")

	regex := regexp.MustCompile(`\n\n+`) // compiling blank lines to 1 blank line // " \s+"
	result = regex.ReplaceAllString(text, "\n\n")
	return
}

// Converting #, ##, * and dates with times
func processLine(line string) string {
	// # 3 ch and ## 4 ch, Dates, Times
	re := regexp.MustCompile(`#([A-Z]{3})|##([A-Z]{4})|D\(([^)]+)\)|T12\(([^)]+)\)|T24\(([^)]+)\)`)
	if bonusFlag { // if bonus
		// *# with 3 characters, *## with 4 characters and so on
		re = regexp.MustCompile(`\*\#([A-Z]{3})|\*\##([A-Z]{4})|#([A-Z]{3})|##([A-Z]{4})|D\(([^)]+)\)|T12\(([^)]+)\)|T24\(([^)]+)\)`)
	} 

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
		return "\033[42m\033[1m\033[37m"+parsedTime.Format("02 Jan 2006")+"\033[0m\033[22m"
	}

	parsedTimeWithOffset, err := time.Parse("D(2006-01-02T15:04-07:00)", isoDate) // if "02:00" formatting
	if err != nil {
		return isoDate
	}

	return "\033[42m\033[1m\033[37m"+parsedTimeWithOffset.Format("02 Jan 2006")+"\033[0m\033[22m"
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

	return fmt.Sprintf("\033[40m\033[32m%s %s\033[0m", formattedTime, offset) // printing human readable
}
