package eidenv

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Go wrapper arounc opensc's eidenv

type Eidenv struct {
	path string
}

type Gender string

const (
	GenderMale   Gender = "M"
	GenderFemale Gender = "F"
	GenderOther  Gender = "X"
)

type DocumentType string

const (
	DocumentTypeEID              = "eID"
	DocumentTypeKidsID           = "Kids ID"
	DocumentTypeBootstrapCard    = "Bootstrap card"
	DocumentTypeHabilitationCard = "Habilitation/machtigings card"
	DocumentTypeACard            = "A-card"
	DocumentTypeBCard            = "B-card"
	DocumentTypeCCard            = "C-card"
	DocumentTypeDCard            = "D-card"
	DocumentTypeECard            = "E-card"
	DocumentTypeEPlusCard        = "E+-card"
	DocumentTypeFCard            = "F-card"
	DocumentTypeFPlusCard        = "F+-card"
	DocumentTypeHCard            = "H-card"
	DocumentTypeICard            = "I-card"
	DocumentTypeJCard            = "J-card"
	DocumentTypeMCard            = "M-card"
	DocumentTypeNCard            = "N-card"
	DocumentTypeKCard            = "K-card"
	DocumentTypeLCard            = "L-card"
	DocumentTypeEUCard           = "EU-card"
	DocumentTypeEUPlusCard       = "EU+-card"
)

var idMap = map[int]DocumentType{
	1:  DocumentTypeEID,
	6:  DocumentTypeKidsID,
	7:  DocumentTypeBootstrapCard,
	8:  DocumentTypeHabilitationCard,
	11: DocumentTypeACard,
	12: DocumentTypeBCard,
	13: DocumentTypeCCard,
	14: DocumentTypeDCard,
	15: DocumentTypeECard,
	16: DocumentTypeEPlusCard,
	17: DocumentTypeFCard,
	18: DocumentTypeFPlusCard,
	19: DocumentTypeHCard,
	20: DocumentTypeICard,
	21: DocumentTypeJCard,
	22: DocumentTypeMCard,
	23: DocumentTypeNCard,
	27: DocumentTypeKCard,
	28: DocumentTypeLCard,
	31: DocumentTypeEUCard,
	32: DocumentTypeEUPlusCard,
	33: DocumentTypeACard,
	34: DocumentTypeBCard,
	35: DocumentTypeFCard,
	36: DocumentTypeFPlusCard,
}

type SpecialStatus string

/* Spec:
0: No status
1: White cane (blind people) (a)
2: Extended minority (a)
3: White cane + extended minority (a)
4: Yellow cane (partially sighted people) (a)
5: Yellow cane + extended minority (a)
*/

const (
	SpecialStatusNoStatus                   SpecialStatus = "No status"
	SpecialStatusWhiteCane                  SpecialStatus = "White cane"
	SpecialStatusExtendedMinority           SpecialStatus = "Extended minority"
	SpecialStatusWhiteCaneExtendedMinority  SpecialStatus = "White cane + extended minority"
	SpecialStatusYellowCane                 SpecialStatus = "Yellow cane"
	SpecialStatusYellowCaneExtendedMinority SpecialStatus = "Yellow cane + extended minority"
)

var specialStatusMap = map[int]SpecialStatus{
	0: SpecialStatusNoStatus,
	1: SpecialStatusWhiteCane,
	2: SpecialStatusExtendedMinority,
	3: SpecialStatusWhiteCaneExtendedMinority,
	4: SpecialStatusYellowCane,
	5: SpecialStatusYellowCaneExtendedMinority,
}

type CardInfo struct {
	CardNumber             string        `json:"cardNumber"`
	ValidFrom              time.Time     `json:"validFrom"`
	ValidTill              time.Time     `json:"validTill"`
	DeliveringMunicipality string        `json:"deliveringMunicipality"`
	NationalNumber         string        `json:"nationalNumber"`
	Name                   string        `json:"name"`
	FirstNames             string        `json:"firstNames"`
	Initial                string        `json:"initial"`
	Nationality            string        `json:"nationality"`
	BirthLocation          string        `json:"birthLocation"`
	BirthDate              time.Time     `json:"birthDate"` // Birth date: DD mmmm YYYY or DD.mmm.YYYY (German, we parse the date
	Gender                 string        `json:"gender"`    // M: man F/V/W: woman, we simplify this to M/F/X
	NobleCondition         string        `json:"nobleCondition"`
	DocumentType           DocumentType  `json:"documentType"`  // TODO: lookup
	SpecialStatus          SpecialStatus `json:"specialStatus"` // TODO: lookup
	Address                string        `json:"address"`
	Zipcode                string        `json:"zipcode"`
	Municipality           string        `json:"municipality"`
}

func New() (*Eidenv, error) {
	// check if eidenv is available in the PATH
	path, err := exec.LookPath("eidenv")
	if err != nil {
		return nil, fmt.Errorf("eidenv not found in PATH: %w", err)
	}
	return &Eidenv{
		path: path,
	}, nil
}

func (e *Eidenv) ReadCard() (CardInfo, error) {
	// exec eidenv
	cmd := exec.Command(e.path)
	stdout, err := cmd.Output()
	if err != nil && !strings.Contains(err.Error(), "exit status 1") { // ignore exit status 1 as for some reason it meant okay...
		return CardInfo{}, fmt.Errorf("failed to read card: %w %q", err, string(stdout))
	}

	// parse output
	cardInfo := CardInfo{}
	scanner := bufio.NewScanner(bytes.NewReader(stdout))

	for scanner.Scan() {
		info := scanner.Text()
		if info == "" {
			continue
		} else if strings.Contains(info, "Card not present") {
			return CardInfo{}, fmt.Errorf("card not present")
		} else if !strings.Contains(info, ":") {
			continue
		}

		parts := strings.SplitN(info, ":", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "BELPIC_CARDNUMBER":
			cardInfo.CardNumber = value
		case "BELPIC_VALIDFROM":
			cardInfo.ValidFrom, _ = time.Parse("02.01.2006", value)
		case "BELPIC_VALIDTILL":
			cardInfo.ValidTill, _ = time.Parse("02.01.2006", value)
		case "BELPIC_DELIVERINGMUNICIPALITY":
			cardInfo.DeliveringMunicipality = value
		case "BELPIC_NATIONALNUMBER":
			cardInfo.NationalNumber = value
		case "BELPIC_NAME":
			cardInfo.Name = value
		case "BELPIC_FIRSTNAMES":
			cardInfo.FirstNames = value
		case "BELPIC_INITIAL":
			cardInfo.Initial = value
		case "BELPIC_NATIONALITY":
			cardInfo.Nationality = value
		case "BELPIC_BIRTHLOCATION":
			cardInfo.BirthLocation = value
		case "BELPIC_BIRTHDATE":
			cardInfo.BirthDate = parseEIDBirthdate(value)
		case "BELPIC_SEX":
			switch value {
			case "M":
				cardInfo.Gender = string(GenderMale)
			case "X":
				cardInfo.Gender = string(GenderOther)
			default:
				cardInfo.Gender = string(GenderFemale) // in EID spec this is V/F/W... quite annoying!
			}
		case "BELPIC_NOBLECONDITION":
			cardInfo.NobleCondition = value
		case "BELPIC_DOCUMENTTYPE":
			id, _ := strconv.ParseInt(value, 10, 32)
			cardInfo.DocumentType = idMap[int(id)]
		case "BELPIC_SPECIALSTATUS":
			id, _ := strconv.ParseInt(value, 10, 32)
			cardInfo.SpecialStatus = specialStatusMap[int(id)]
		case "BELPIC_STREETANDNUMBER":
			cardInfo.Address = value
		case "BELPIC_ZIPCODE":
			cardInfo.Zipcode = value
		case "BELPIC_MUNICIPALITY":
			cardInfo.Municipality = value
		}

	}

	return cardInfo, nil
}
