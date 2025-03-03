package test

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"
)

const lookupHeader = "name,iso_country,municipality,icao_code,iata_code,coordinates"
const lookupBasicBody = `
Honiara International Airport,SB,Honiara,AGGH,HIR,"160.05499267578, -9.4280004501343"
Hongyuan Airport,CN,Aba,ZUHY,AHJ,"102.35224, 32.53154"
Nauru International Airport,NR,Yaren District,ANYN,INU,"166.919006, -0.547458"
Alxa Left Banner Bayanhot Airport,CN,Bayanhot,ZBAL,AXF,"105.58858, 38.74831"
Buka Airport,PG,Buka Island,AYBK,BUA,"154.67300415039062, -5.4223198890686035"`
const basicLookup = lookupHeader + lookupBasicBody

type airport struct {
	name string
	city string
	icao string
	iata string
}

func TestEmptyInput(t *testing.T) {
	runWithMockFiles(t, "", basicLookup, "", true, 0)
}

func TestValidAirportCodes(t *testing.T) {
	data := []airport{
		{"Honiara International Airport", "Honiara", "AGGH", "HIR"},
		{"Hongyuan Airport", "Aba", "ZUHY", "AHJ"},
		{"Nauru International Airport", "Yaren District", "ANYN", "INU"},
		{"Alxa Left Banner Bayanhot Airport", "Bayanhot", "ZBAL", "AXF"},
		{"Buka Airport", "Buka Island", "AYBK", "BUA"},
	}

	t.Run("NoHash", func(t *testing.T) {
		var input strings.Builder
		var output strings.Builder

		for i, ap := range data {
			input.WriteString(ap.iata + "\n")
			input.WriteString(ap.icao)
			output.WriteString(ap.iata + "\n")
			output.WriteString(ap.icao)

			if i+1 < len(data) {
				input.WriteByte('\n')
				output.WriteByte('\n')
			}
		}

		runWithMockFiles(t, input.String(), basicLookup, output.String(), false, 5)
	})

	t.Run("IATA", func(t *testing.T) {
		var input strings.Builder
		var output strings.Builder

		for i, ap := range data {
			input.WriteString("#" + ap.iata)
			output.WriteString(ap.name)
			if i+1 < len(data) {
				input.WriteByte('\n')
				output.WriteByte('\n')
			}
		}

		runWithMockFiles(t, input.String(), basicLookup, output.String(), false, 5)
	})

	t.Run("ICAO", func(t *testing.T) {
		var input strings.Builder
		var output strings.Builder

		for i, ap := range data {
			input.WriteString("##" + ap.icao)
			output.WriteString(ap.name)
			if i+1 < len(data) {
				input.WriteByte('\n')
				output.WriteByte('\n')
			}
		}

		runWithMockFiles(t, input.String(), basicLookup, output.String(), false, 5)
	})

	t.Run("InText", func(t *testing.T) {
		var input strings.Builder
		var output strings.Builder

		format := "Departing from %s | %s on the %d day"
		for i, ap := range data {
			input.WriteString(fmt.Sprintf(format, "#"+ap.iata, "##"+ap.icao, i))
			output.WriteString(fmt.Sprintf(format, ap.name, ap.name, i))
			if i+1 < len(data) {
				input.WriteByte('\n')
				output.WriteByte('\n')
			}
		}

		runWithMockFiles(t, input.String(), basicLookup, output.String(), false, 5)
	})

	t.Run("FromReview", func(t *testing.T) {
		const input = "Your flight departs from #INU, and your destination is ##AYBK."
		const expected = "Your flight departs from Nauru International Airport, and your destination is Buka Airport."

		runWithMockFiles(t, input, basicLookup, expected, false, 5)
	})
}

func TestAiportCodesSurroundedWithWordBreak(t *testing.T) {
	/*
		space          #HIR   ##AGGH   *#HIR   *##AGGH  = Honiara International Airport,     Honiara
		parens        (#AHJ) (##ZUHY) (*#AHJ) (*##ZUHY) = Hongyuan Airport,                  Aba
		brackets      [#INU] [##ANYN] [*#INU] [*##ANYN] = Nauru International Airport,       Yaren District
		braces        {#AXF} {##ZBAL} {*#AXF} {*##ZBAL} = Alxa Left Banner Bayanhot Airport, Bayanhot
		chevrons      <#BUA> <##AYBK> <*#BUA> <*##AYBK> = Buka Airport,                      Buka Island
		commas        ,#HIR, ,##AGGH, ,*#HIR, ,*##AGGH, = Honiara International Airport,     Honiara
		dots          .#AHJ. .##ZUHY. .*#AHJ. .*##ZUHY. = Hongyuan Airport,                  Aba
		slash         /#INU/ /##ANYN/ /*#INU/ /*##ANYN/ = Nauru International Airport,       Yaren District
		backslash     \#AXF\ \##ZBAL\ \*#AXF\ \*##ZBAL\ = Alxa Left Banner Bayanhot Airport, Bayanhot
		pipe          |#BUA| |##AYBK| |*#BUA| |*##AYBK| = Buka Airport,                      Buka Island
		caret         ^#HIR^ ^##AGGH^ ^*#HIR^ ^*##AGGH^ = Honiara International Airport,     Honiara
		colon         :#AHJ: :##ZUHY: :*#AHJ: :*##ZUHY: = Hongyuan Airport,                  Aba
		semicolon     ;#INU; ;##ANYN; ;*#INU; ;*##ANYN; = Nauru International Airport,       Yaren District
		double-quotes "#AXF" "##ZBAL" "*#AXF" "*##ZBAL" = Alxa Left Banner Bayanhot Airport, Bayanhot
		single-quotes '#BUA' '##AYBK' '*#BUA' '*##AYBK' = Buka Airport,                      Buka Island
		backtick      `#HIR` `##AGGH` `*#HIR` `*##AGGH` = Honiara International Airport,     Honiara
	*/

	cases := [][]string{
		// Airports
		{"space airport", "a #HIR ##AGGH a", "a Honiara International Airport Honiara International Airport a"},
		{"parens (airport)", `(#AHJ) (##ZUHY)`, `(Hongyuan Airport) (Hongyuan Airport)`},
		{"brackets [airport]", `[#INU] [##ANYN]`, `[Nauru International Airport] [Nauru International Airport]`},
		{"braces {airport}", `{#AXF} {##ZBAL}`, `{Alxa Left Banner Bayanhot Airport} {Alxa Left Banner Bayanhot Airport}`},
		{"chevrons <airport>", `<#BUA> <##AYBK>`, `<Buka Airport> <Buka Airport>`},
		{"commas ,airport,", `,#HIR, ,##AGGH,`, `,Honiara International Airport, ,Honiara International Airport,`},
		{"dots .airport.", `.#AHJ. .##ZUHY.`, `.Hongyuan Airport. .Hongyuan Airport.`},
		{"slash /airport/", `/#INU/ /##ANYN/`, `/Nauru International Airport/ /Nauru International Airport/`},
		{"backslash \\airport\\", `\#AXF\ \##ZBAL\`, `\Alxa Left Banner Bayanhot Airport\ \Alxa Left Banner Bayanhot Airport\`},
		{"pipe |airport|", `|#BUA| |##AYBK|`, `|Buka Airport| |Buka Airport|`},
		{"caret ^airport^", `^#HIR^ ^##AGGH^`, `^Honiara International Airport^ ^Honiara International Airport^`},
		{"colon :airport:", `:#AHJ: :##ZUHY:`, `:Hongyuan Airport: :Hongyuan Airport:`},
		{"semicolon ;airport;", `;#INU; ;##ANYN;`, `;Nauru International Airport; ;Nauru International Airport;`},
		{"double-quotes \"airport\"", `"#AXF" "##ZBAL"`, `"Alxa Left Banner Bayanhot Airport" "Alxa Left Banner Bayanhot Airport"`},
		{"single-quotes 'airport'", `'#BUA' '##AYBK'`, `'Buka Airport' 'Buka Airport'`},
		{"backtick `airport`", "`#HIR` `##AGGH`", "`Honiara International Airport` `Honiara International Airport`"},
		// Cities
		{"space city", "a *#HIR *##AGGH a", "a Honiara Honiara a"},
		{"parens (city)", `(*#AHJ) (*##ZUHY)`, `(Aba) (Aba)`},
		{"brackets [city]", `[*#INU] [*##ANYN]`, `[Yaren District] [Yaren District]`},
		{"braces {city}", `{*#AXF} {*##ZBAL}`, `{Bayanhot} {Bayanhot}`},
		{"chevrons <city>", `<*#BUA> <*##AYBK>`, `<Buka Island> <Buka Island>`},
		{"commas ,city,", `,*#HIR, ,*##AGGH,`, `,Honiara, ,Honiara,`},
		{"dots .city.", `.*#AHJ. .*##ZUHY.`, `.Aba. .Aba.`},
		{"slash /city/", `/*#INU/ /*##ANYN/`, `/Yaren District/ /Yaren District/`},
		{"backslash \\city\\", `\*#AXF\ \*##ZBAL\`, `\Bayanhot\ \Bayanhot\`},
		{"pipe |city|", `|*#BUA| |*##AYBK|`, `|Buka Island| |Buka Island|`},
		{"caret ^city^", `^*#HIR^ ^*##AGGH^`, `^Honiara^ ^Honiara^`},
		{"colon :city:", `:*#AHJ: :*##ZUHY:`, `:Aba: :Aba:`},
		{"semicolon ;city;", `;*#INU; ;*##ANYN;`, `;Yaren District; ;Yaren District;`},
		{"double-quotes \"city\"", `"*#AXF" "*##ZBAL"`, `"Bayanhot" "Bayanhot"`},
		{"single-quotes 'city'", `'*#BUA' '*##AYBK'`, `'Buka Island' 'Buka Island'`},
		{"backtick `city`", "`*#HIR` `*##AGGH`", "`Honiara` `Honiara`"},
	}

	for _, c := range cases {
		name, input, expected := c[0], c[1], c[2]
		t.Run(name, func(t *testing.T) {
			runWithMockFiles(t, input, basicLookup, expected, false, 5)
		})
	}
}

func TestNonConvertibleAirportCodes(t *testing.T) {
	t.Run("SurroundedWithAlphanum", func(t *testing.T) {
		const input = `word#AHJ word##ZUHY word*AHJ word*##ZUHY
#AHJword ##ZUHYword *#AHJword *##ZUHYword
word#AHJword word##ZUHYword word*AHJword word*##ZUHYword
123#BUA 123##AYBK 123*#BUA 123*##AYBK
#BUA123 ##AYBK123 *#BUA123 *##AYBK123
123#BUA123 123##AYBK123 123*#BUA123 123*##AYBK123`
		const expected = input

		runWithMockFiles(t, input, basicLookup, expected, false, 5)
	})

	t.Run("Lowercase", func(t *testing.T) {
		// It's expected to not fail nor convert anything
		const input = "#abc\n##defg\n*#hij\n*##klmn"
		const expected = input

		runWithMockFiles(t, input, basicLookup, expected, false, 5)
	})

	t.Run("NumbersInCodes", func(t *testing.T) {
		const input = "#A1B\n##C23D\n*#E4F\n*##G56H"
		const expected = input

		runWithMockFiles(t, input, basicLookup, expected, false, 5)
	})
}

func TestValidAirportCities(t *testing.T) {
	data := []airport{
		{"Honiara International Airport", "Honiara", "AGGH", "HIR"},
		{"Hongyuan Airport", "Aba", "ZUHY", "AHJ"},
		{"Nauru International Airport", "Yaren District", "ANYN", "INU"},
		{"Alxa Left Banner Bayanhot Airport", "Bayanhot", "ZBAL", "AXF"},
		{"Buka Airport", "Buka Island", "AYBK", "BUA"},
	}

	t.Run("NoHash", func(t *testing.T) {
		var input strings.Builder
		var output strings.Builder

		for i, ap := range data {
			input.WriteString("*" + ap.iata + "\n")
			input.WriteString("*" + ap.icao)
			output.WriteString("*" + ap.iata + "\n")
			output.WriteString("*" + ap.icao)

			if i+1 < len(data) {
				input.WriteByte('\n')
				output.WriteByte('\n')
			}
		}

		runWithMockFiles(t, input.String(), basicLookup, output.String(), false, 5)
	})

	t.Run("IATA", func(t *testing.T) {
		var input strings.Builder
		var output strings.Builder

		for i, ap := range data {
			input.WriteString("*#" + ap.iata)
			output.WriteString(ap.city)
			if i+1 < len(data) {
				input.WriteByte('\n')
				output.WriteByte('\n')
			}
		}

		runWithMockFiles(t, input.String(), basicLookup, output.String(), false, 5)
	})

	t.Run("ICAO", func(t *testing.T) {
		var input strings.Builder
		var output strings.Builder

		for i, ap := range data {
			input.WriteString("*##" + ap.icao)
			output.WriteString(ap.city)
			if i+1 < len(data) {
				input.WriteByte('\n')
				output.WriteByte('\n')
			}
		}

		runWithMockFiles(t, input.String(), basicLookup, output.String(), false, 5)
	})

	t.Run("InText", func(t *testing.T) {
		var input strings.Builder
		var output strings.Builder

		format := "Departing from %s | %s on the %d day"
		for i, ap := range data {
			input.WriteString(fmt.Sprintf(format, "*#"+ap.iata, "*##"+ap.icao, i))
			output.WriteString(fmt.Sprintf(format, ap.city, ap.city, i))
			if i+1 < len(data) {
				input.WriteByte('\n')
				output.WriteByte('\n')
			}
		}

		runWithMockFiles(t, input.String(), basicLookup, output.String(), false, 5)
	})
}

func TestValidDates(t *testing.T) {
	data := [][]string{
		{"D(2024-02-01T08:00-08:00)", "01 Feb 2024"},
		{"T12(2024-02-01T08:00-08:00)", "08:00AM (-08:00)"},
		{"T24(2024-02-01T08:00-08:00)", "08:00 (-08:00)"},
		{"D(2024-02-01T16:30-05:00)", "01 Feb 2024"},
		{"T12(2024-02-01T16:30-05:00)", "04:30PM (-05:00)"},
		{"T24(2024-02-01T16:30-05:00)", "16:30 (-05:00)"},
		{"D(2024-02-02T09:45-05:00)", "02 Feb 2024"},
		{"T12(2024-02-02T09:45-05:00)", "09:45AM (-05:00)"},
		{"T24(2024-02-02T09:45-05:00)", "09:45 (-05:00)"},
		{"D(2024-02-02T22:00+00:00)", "02 Feb 2024"},
		{"T12(2024-02-02T22:00+00:00)", "10:00PM (+00:00)"},
		{"T24(2024-02-02T22:00+00:00)", "22:00 (+00:00)"},
		{"D(2024-02-03T08:30+01:00)", "03 Feb 2024"},
		{"T12(2024-02-03T08:30+01:00)", "08:30AM (+01:00)"},
		{"T24(2024-02-03T08:30+01:00)", "08:30 (+01:00)"},
		{"D(2024-02-03T14:15+04:00)", "03 Feb 2024"},
		{"T12(2024-02-03T14:15+04:00)", "02:15PM (+04:00)"},
		{"T24(2024-02-03T14:15+04:00)", "14:15 (+04:00)"},
		{"D(2024-02-03T23:45+04:00)", "03 Feb 2024"},
		{"T12(2024-02-03T23:45+04:00)", "11:45PM (+04:00)"},
		{"T24(2024-02-03T23:45+04:00)", "23:45 (+04:00)"},
		{"D(2024-02-04T03:30+08:00)", "04 Feb 2024"},
		{"T12(2024-02-04T03:30+08:00)", "03:30AM (+08:00)"},
		{"T24(2024-02-04T03:30+08:00)", "03:30 (+08:00)"},
		{"D(2024-02-04T15:00+08:00)", "04 Feb 2024"},
		{"T12(2024-02-04T15:00+08:00)", "03:00PM (+08:00)"},
		{"T24(2024-02-04T15:00+08:00)", "15:00 (+08:00)"},
		{"D(2024-02-05T09:15+11:00)", "05 Feb 2024"},
		{"T12(2024-02-05T09:15+11:00)", "09:15AM (+11:00)"},
		{"T24(2024-02-05T09:15+11:00)", "09:15 (+11:00)"},
		{"D(2024-02-05T21:30+11:00)", "05 Feb 2024"},
		{"T12(2024-02-05T21:30+11:00)", "09:30PM (+11:00)"},
		{"T24(2024-02-05T21:30+11:00)", "21:30 (+11:00)"},
		{"D(2024-02-06T05:45+09:00)", "06 Feb 2024"},
		{"T12(2024-02-06T05:45+09:00)", "05:45AM (+09:00)"},
		{"T24(2024-02-06T05:45+09:00)", "05:45 (+09:00)"},
		{"D(2024-02-06T15:00+09:00)", "06 Feb 2024"},
		{"T12(2024-02-06T15:00+09:00)", "03:00PM (+09:00)"},
		{"T24(2024-02-06T15:00+09:00)", "15:00 (+09:00)"},
		{"D(2024-02-07T07:30+08:00)", "07 Feb 2024"},
		{"T12(2024-02-07T07:30+08:00)", "07:30AM (+08:00)"},
		{"T24(2024-02-07T07:30+08:00)", "07:30 (+08:00)"},
		{"D(2024-02-07T13:45+08:00)", "07 Feb 2024"},
		{"T12(2024-02-07T13:45+08:00)", "01:45PM (+08:00)"},
		{"T24(2024-02-07T13:45+08:00)", "13:45 (+08:00)"},
		{"D(2024-02-08T01:00+09:00)", "08 Feb 2024"},
		{"T12(2024-02-08T01:00+09:00)", "01:00AM (+09:00)"},
		{"T24(2024-02-08T01:00+09:00)", "01:00 (+09:00)"},
		{"D(2024-02-08T06:30+09:00)", "08 Feb 2024"},
		{"T12(2024-02-08T06:30+09:00)", "06:30AM (+09:00)"},
		{"T24(2024-02-08T06:30+09:00)", "06:30 (+09:00)"},
		{"D(2024-02-08T10:15+09:00)", "08 Feb 2024"},
		{"T12(2024-02-08T10:15+09:00)", "10:15AM (+09:00)"},
		{"T24(2024-02-08T10:15+09:00)", "10:15 (+09:00)"},
		{"D(2024-02-08T06:30-08:00)", "08 Feb 2024"},
		{"T12(2024-02-08T06:30-08:00)", "06:30AM (-08:00)"},
		{"T24(2024-02-08T06:30-08:00)", "06:30 (-08:00)"},
		{"D(2024-03-01T10:00+01:00)", "01 Mar 2024"},
		{"T12(2024-03-01T10:00+01:00)", "10:00AM (+01:00)"},
		{"T24(2024-03-01T10:00+01:00)", "10:00 (+01:00)"},
		{"D(2024-03-01T14:30+02:00)", "01 Mar 2024"},
		{"T12(2024-03-01T14:30+02:00)", "02:30PM (+02:00)"},
		{"T24(2024-03-01T14:30+02:00)", "14:30 (+02:00)"},
		{"D(2024-03-02T08:45+02:00)", "02 Mar 2024"},
		{"T12(2024-03-02T08:45+02:00)", "08:45AM (+02:00)"},
		{"T24(2024-03-02T08:45+02:00)", "08:45 (+02:00)"},
		{"D(2024-03-02T12:00+03:00)", "02 Mar 2024"},
		{"T12(2024-03-02T12:00+03:00)", "12:00PM (+03:00)"},
		{"T24(2024-03-02T12:00+03:00)", "12:00 (+03:00)"},
		{"D(2024-03-03T14:30+04:00)", "03 Mar 2024"},
		{"T12(2024-03-03T14:30+04:00)", "02:30PM (+04:00)"},
		{"T24(2024-03-03T14:30+04:00)", "14:30 (+04:00)"},
		{"D(2024-03-03T23:00+04:00)", "03 Mar 2024"},
		{"T12(2024-03-03T23:00+04:00)", "11:00PM (+04:00)"},
		{"T24(2024-03-03T23:00+04:00)", "23:00 (+04:00)"},
		{"D(2024-03-04T08:00+08:00)", "04 Mar 2024"},
		{"T12(2024-03-04T08:00+08:00)", "08:00AM (+08:00)"},
		{"T24(2024-03-04T08:00+08:00)", "08:00 (+08:00)"},
		{"D(2024-03-04T14:30+08:00)", "04 Mar 2024"},
		{"T12(2024-03-04T14:30+08:00)", "02:30PM (+08:00)"},
		{"T24(2024-03-04T14:30+08:00)", "14:30 (+08:00)"},
		{"D(2024-03-05T02:00+08:00)", "05 Mar 2024"},
		{"T12(2024-03-05T02:00+08:00)", "02:00AM (+08:00)"},
		{"T24(2024-03-05T02:00+08:00)", "02:00 (+08:00)"},
		{"D(2024-03-05T15:30+08:00)", "05 Mar 2024"},
		{"T12(2024-03-05T15:30+08:00)", "03:30PM (+08:00)"},
		{"T24(2024-03-05T15:30+08:00)", "15:30 (+08:00)"},
		{"D(2024-03-06T09:30+05:00)", "06 Mar 2024"},
		{"T12(2024-03-06T09:30+05:00)", "09:30AM (+05:00)"},
		{"T24(2024-03-06T09:30+05:00)", "09:30 (+05:00)"},
		{"D(2024-03-06T18:00+05:00)", "06 Mar 2024"},
		{"T12(2024-03-06T18:00+05:00)", "06:00PM (+05:00)"},
		{"T24(2024-03-06T18:00+05:00)", "18:00 (+05:00)"},
		{"D(2024-03-07T04:15+08:00)", "07 Mar 2024"},
		{"T12(2024-03-07T04:15+08:00)", "04:15AM (+08:00)"},
		{"T24(2024-03-07T04:15+08:00)", "04:15 (+08:00)"},
		{"D(2024-03-07T09:30+08:00)", "07 Mar 2024"},
		{"T12(2024-03-07T09:30+08:00)", "09:30AM (+08:00)"},
		{"T24(2024-03-07T09:30+08:00)", "09:30 (+08:00)"},
	}

	t.Run("Review", func(t *testing.T) {
		data := [][]string{
			{"D(2022-05-09T08:07Z)", "09 May 2022"},
			{"T12(2069-04-24T19:18-02:00)", "07:18PM (-02:00)"},
			{"T12(2080-05-04T14:54Z)", "02:54PM (+00:00)"},
			{"T12(1980-02-17T03:30+11:00)", "03:30AM (+11:00)"},
			{"T12(2029-09-04T03:09Z)", "03:09AM (+00:00)"},
			{"T24(2032-07-17T04:08+13:00)", "04:08 (+13:00)"},
			{"T24(2084-04-13T17:54Z)", "17:54 (+00:00)"},
			{"T24(2024-07-23T15:29-11:00)", "15:29 (-11:00)"},
			{"T24(2042-09-01T21:43Z)", "21:43 (+00:00)"},
		}

		var input strings.Builder
		var output strings.Builder

		for i, d := range data {
			input.WriteString(d[0])
			output.WriteString(d[1])
			if i+1 < len(data) {
				input.WriteByte('\n')
				output.WriteByte('\n')
			}
		}

		runWithMockFiles(t, input.String(), basicLookup, output.String(), false, 5)
	})

	t.Run("Long", func(t *testing.T) {
		var input strings.Builder
		var output strings.Builder

		for i, d := range data {
			input.WriteString(d[0])
			output.WriteString(d[1])
			if i+1 < len(data) {
				input.WriteByte('\n')
				output.WriteByte('\n')
			}
		}

		runWithMockFiles(t, input.String(), basicLookup, output.String(), false, 5)
	})

	t.Run("InText", func(t *testing.T) {
		var input strings.Builder
		var output strings.Builder

		format := "Departing from A-%s ( %s ) to B-%s"
		for i, d := range data {
			a := strconv.FormatInt(int64(i), 36)
			input.WriteString(fmt.Sprintf(format, a, d[0], a))
			output.WriteString(fmt.Sprintf(format, a, d[1], a))
			if i+1 < len(data) {
				input.WriteByte('\n')
				output.WriteByte('\n')
			}
		}

		runWithMockFiles(t, input.String(), basicLookup, output.String(), false, 5)
	})
}

func TestRandomDates(t *testing.T) {
	const amount = 1000

	var data = make([]time.Time, 0, amount)

	for i := 0; i < amount; i++ {
		data = append(data, getRandomDatetime())
	}

	t.Run("Regular", func(t *testing.T) {
		var input strings.Builder
		var output strings.Builder

		for i, dt := range data {
			iso, d, t12, t24 := getDatetimeTestStrings(dt)
			input.WriteString("D(" + iso + ")\n")
			input.WriteString("T12(" + iso + ")\n")
			input.WriteString("T24(" + iso + ")")

			output.WriteString(d + "\n")
			output.WriteString(t12 + "\n")
			output.WriteString(t24)

			if i+1 < len(data) {
				input.WriteByte('\n')
				output.WriteByte('\n')
			}
		}

		runWithMockFiles(t, input.String(), basicLookup, output.String(), false, 5)
	})

	t.Run("InText", func(t *testing.T) {
		var input strings.Builder
		var output strings.Builder

		var (
			formatD   = "Departure on %s from A-%s\n"
			formatT12 = "Layover at %s in B-%s\n"
			formatT24 = "Arrival to %s to C-%s"
		)

		for i, dt := range data {
			a := strconv.FormatInt(int64(i), 36)
			iso, d, t12, t24 := getDatetimeTestStrings(dt)
			input.WriteString(fmt.Sprintf(formatD, "D("+iso+")", a))
			input.WriteString(fmt.Sprintf(formatT12, "T12("+iso+")", a))
			input.WriteString(fmt.Sprintf(formatT24, "T24("+iso+")", a))

			output.WriteString(fmt.Sprintf(formatD, d, a))
			output.WriteString(fmt.Sprintf(formatT12, t12, a))
			output.WriteString(fmt.Sprintf(formatT24, t24, a))

			if i+1 < len(data) {
				input.WriteByte('\n')
				output.WriteByte('\n')
			}
		}

		runWithMockFiles(t, input.String(), basicLookup, output.String(), false, 5)
	})
}

func TestSpecialCharsConvertion(t *testing.T) {
	const input = "Lorem ipsum dolor sit amet,\vconsectetur adipiscing elit. \nProin risus nisi, \fcongue ut tempor ac, \rlacinia ac justo. \rQuisque sed felis vestibulum, \vcommodo lacus quis, \faliquam libero. Integer vitae \fpellentesque dolor. \nCurabitur finibus sapien et diam interdum, a vulputate ex euismod. \vSed mattis, tortor vel lacinia luctus, diam risus tempus purus, \rvitae fermentum arcu massa non purus. \rNam nulla ex, pellentesque quis ligula quis, \vfermentum ullamcorper urna. \fMaecenas dapibus consectetur elit. \vQuisque mattis rhoncus lacinia. In a fringilla tellus, in aliquam est. \rIn hac habitasse platea dictumst. Nulla imperdiet arcu sed auctor \fornare. Proin quis eros a erat imperdiet pretium. Aliquam \rpretium mollis purus \veu consectetur."

	const output = "Lorem ipsum dolor sit amet,\nconsectetur adipiscing elit. \nProin risus nisi, \ncongue ut tempor ac, \nlacinia ac justo. \nQuisque sed felis vestibulum, \ncommodo lacus quis, \naliquam libero. Integer vitae \npellentesque dolor. \nCurabitur finibus sapien et diam interdum, a vulputate ex euismod. \nSed mattis, tortor vel lacinia luctus, diam risus tempus purus, \nvitae fermentum arcu massa non purus. \nNam nulla ex, pellentesque quis ligula quis, \nfermentum ullamcorper urna. \nMaecenas dapibus consectetur elit. \nQuisque mattis rhoncus lacinia. In a fringilla tellus, in aliquam est. \nIn hac habitasse platea dictumst. Nulla imperdiet arcu sed auctor \nornare. Proin quis eros a erat imperdiet pretium. Aliquam \npretium mollis purus \neu consectetur."

	runWithMockFiles(t, input, basicLookup, output, false, 10)
}

func TestMultipleNewlines(t *testing.T) {
	const input = "A\nB\n\nC\n\n\nD\n\n\n\nE\n\n\n\n\nF"
	const output = "A\nB\n\nC\n\nD\n\nE\n\nF"

	runWithMockFiles(t, input, basicLookup, output, false, 5)
}

func TestExcessiveSpace(t *testing.T) {
	const input = "A B  C   D    E      F               G\n    H   I    \nJ K"
	const expected = "A B C D E F G\nH I\nJ K"

	runWithMockFiles(t, input, basicLookup, expected, false, 5)
}

func TestSpecialCharsRepeated(t *testing.T) {
	const specialChars = "\v\f\r"
	const scl = len(specialChars)
	const targetChar byte = '\n'

	var input strings.Builder
	var output strings.Builder

	var n int64 = 0

	wn := func() {
		n++
		ns := strconv.FormatInt(n, 36)
		input.WriteString(ns)
		output.WriteString(ns)
	}

	// 1 special char in a row
	for i := 0; i < scl; i++ {
		wn()
		input.WriteByte(specialChars[i])
		output.WriteByte(targetChar)
	}

	// 2 special chars in a row
	for i := 0; i < scl; i++ {
		for j := 0; j < scl; j++ {
			wn()
			input.WriteByte(specialChars[i])
			input.WriteByte(specialChars[j])
			output.WriteByte(targetChar)
			output.WriteByte(targetChar)
		}
	}

	// 3 special chars in a row
	for i := 0; i < scl; i++ {
		for j := 0; j < scl; j++ {
			for k := 0; k < scl; k++ {
				wn()
				input.WriteByte(specialChars[i])
				input.WriteByte(specialChars[j])
				input.WriteByte(specialChars[k])
				output.WriteByte(targetChar)
				output.WriteByte(targetChar)
			}
		}
	}

	// 4 special chars in a row
	for i := 0; i < scl; i++ {
		for j := 0; j < scl; j++ {
			for k := 0; k < scl; k++ {
				for l := 0; l < scl; l++ {
					n++
					ns := strconv.FormatInt(n, 36)
					input.WriteString(ns)
					output.WriteString(ns)
					input.WriteByte(specialChars[i])
					input.WriteByte(specialChars[j])
					input.WriteByte(specialChars[k])
					input.WriteByte(specialChars[l])
					output.WriteByte(targetChar)
					output.WriteByte(targetChar)
				}
			}
		}
	}

	input.WriteRune('$')
	output.WriteRune('$')

	runWithMockFiles(t, input.String(), basicLookup, output.String(), true, 10)
}

func getRandomDatetime() time.Time {
	var (
		year     = 1900 + rand.Intn(1000)
		month    = 1 + rand.Intn(12)
		day      = 1 + rand.Intn(25)
		hour     = rand.Intn(24)
		minute   = rand.Intn(60)
		timezone = -11 + rand.Intn(24)
	)

	t, err := time.Parse(time.RFC3339, fmt.Sprintf("%02d-%02d-%02dT%02d:%02d:00%+03d:00", year, month, day, hour, minute, timezone))
	if err != nil {
		panic(err)
	}
	return t
}

// getDatetimeTestStrings formats time to test strings
// first string is ISO 8601 test string that is used in input
// second string is ISO 8601 with Z test string that is used in input
// second string is expected D(date)
// third string is expected T12(date)
// fourth string is expected T24(date)
func getDatetimeTestStrings(dt time.Time) (string, string, string, string) {
	iso := dt.Format("2006-01-02T15:04-07:00")
	d := dt.Format("02 Jan 2006")
	t12 := dt.Format("03:04PM (-07:00)")
	t24 := dt.Format("15:04 (-07:00)")

	return iso, d, t12, t24
}
