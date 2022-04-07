package eidenv

import (
	"strings"
	"time"
)

/* Months specified in spec
JAN FEV MARS AVR MAI JUIN JUIL AOUT SEPT OCT NOV DEC
JAN FEB MAAR APR MEI JUN JUL AUG SEP OKT NOV DEC
JAN FEB MÄR APR MAI JUN JUL AUG SEP OKT NOV DEZ
*/

var monthTranslations = map[string]string{
	"JAN":  "01",
	"FEB":  "02",
	"FEV":  "02",
	"MÄR":  "03",
	"MAAR": "03",
	"MARS": "03",
	"APR":  "04",
	"AVR":  "04",
	"MAI":  "05",
	"MEI":  "05",
	"JUN":  "06",
	"JUIN": "06",
	"JUL":  "07",
	"JUIL": "07",
	"AUG":  "08",
	"AOUT": "08",
	"SEP":  "09",
	"SEPT": "09",
	"OKT":  "10",
	"OCT":  "10",
	"NOV":  "11",
	"DEZ":  "12",
	"DEC":  "12",
}

func parseEIDBirthdate(in string) time.Time {
	/* spec:
	Birth date: DD mmmm YYYY r DD.mmm.YYYY (German
	may also contain spaces too much for reasons...
	*/
	parts := leaveOutEmpty(strings.Split(in, " "))
	if len(parts) != 3 {
		// german date
		for month, num := range monthTranslations {
			in = strings.Replace(in, month, num, -1)

			t, _ := time.Parse("02.01.2006", in)
			return t
		}
	}

	for month, num := range monthTranslations {
		parts[1] = strings.Replace(parts[1], month, num, -1)
	}

	t, _ := time.Parse("02.01.2006", strings.Join(parts, "."))

	return t
}

func leaveOutEmpty(in []string) []string {
	out := []string{}
	for _, part := range in {
		if strings.TrimSpace(part) != "" {
			out = append(out, part)
		}
	}
	return out
}
