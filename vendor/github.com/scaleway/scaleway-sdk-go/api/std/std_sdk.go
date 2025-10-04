// This file was automatically generated. DO NOT EDIT.
// If you have any remark or suggestion do not hesitate to open an issue.

// Package std provides methods and message types of the std  API.
package std

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/scaleway/scaleway-sdk-go/marshaler"
	"github.com/scaleway/scaleway-sdk-go/namegenerator"
	"github.com/scaleway/scaleway-sdk-go/parameter"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// always import dependencies
var (
	_ fmt.Stringer
	_ json.Unmarshaler
	_ url.URL
	_ net.IP
	_ http.Header
	_ bytes.Reader
	_ time.Time
	_ = strings.Join

	_ scw.ScalewayRequest
	_ marshaler.Duration
	_ scw.File
	_ = parameter.AddToQuery
	_ = namegenerator.GetRandomName
)

type CountryCode string

const (
	CountryCodeUnknownCountryCode = CountryCode("unknown_country_code")
	CountryCodeAF                 = CountryCode("AF")
	CountryCodeAX                 = CountryCode("AX")
	CountryCodeAL                 = CountryCode("AL")
	CountryCodeDZ                 = CountryCode("DZ")
	CountryCodeAS                 = CountryCode("AS")
	CountryCodeAD                 = CountryCode("AD")
	CountryCodeAO                 = CountryCode("AO")
	CountryCodeAI                 = CountryCode("AI")
	CountryCodeAQ                 = CountryCode("AQ")
	CountryCodeAG                 = CountryCode("AG")
	CountryCodeAR                 = CountryCode("AR")
	CountryCodeAM                 = CountryCode("AM")
	CountryCodeAW                 = CountryCode("AW")
	CountryCodeAU                 = CountryCode("AU")
	CountryCodeAT                 = CountryCode("AT")
	CountryCodeAZ                 = CountryCode("AZ")
	CountryCodeBS                 = CountryCode("BS")
	CountryCodeBH                 = CountryCode("BH")
	CountryCodeBD                 = CountryCode("BD")
	CountryCodeBB                 = CountryCode("BB")
	CountryCodeBY                 = CountryCode("BY")
	CountryCodeBE                 = CountryCode("BE")
	CountryCodeBZ                 = CountryCode("BZ")
	CountryCodeBJ                 = CountryCode("BJ")
	CountryCodeBM                 = CountryCode("BM")
	CountryCodeBT                 = CountryCode("BT")
	CountryCodeBO                 = CountryCode("BO")
	CountryCodeBQ                 = CountryCode("BQ")
	CountryCodeBA                 = CountryCode("BA")
	CountryCodeBW                 = CountryCode("BW")
	CountryCodeBV                 = CountryCode("BV")
	CountryCodeBR                 = CountryCode("BR")
	CountryCodeIO                 = CountryCode("IO")
	CountryCodeBN                 = CountryCode("BN")
	CountryCodeBG                 = CountryCode("BG")
	CountryCodeBF                 = CountryCode("BF")
	CountryCodeBI                 = CountryCode("BI")
	CountryCodeKH                 = CountryCode("KH")
	CountryCodeCM                 = CountryCode("CM")
	CountryCodeCA                 = CountryCode("CA")
	CountryCodeCV                 = CountryCode("CV")
	CountryCodeKY                 = CountryCode("KY")
	CountryCodeCF                 = CountryCode("CF")
	CountryCodeTD                 = CountryCode("TD")
	CountryCodeCL                 = CountryCode("CL")
	CountryCodeCN                 = CountryCode("CN")
	CountryCodeCX                 = CountryCode("CX")
	CountryCodeCC                 = CountryCode("CC")
	CountryCodeCO                 = CountryCode("CO")
	CountryCodeKM                 = CountryCode("KM")
	CountryCodeCG                 = CountryCode("CG")
	CountryCodeCD                 = CountryCode("CD")
	CountryCodeCK                 = CountryCode("CK")
	CountryCodeCR                 = CountryCode("CR")
	CountryCodeCI                 = CountryCode("CI")
	CountryCodeHR                 = CountryCode("HR")
	CountryCodeCU                 = CountryCode("CU")
	CountryCodeCW                 = CountryCode("CW")
	CountryCodeCY                 = CountryCode("CY")
	CountryCodeCZ                 = CountryCode("CZ")
	CountryCodeDK                 = CountryCode("DK")
	CountryCodeDJ                 = CountryCode("DJ")
	CountryCodeDM                 = CountryCode("DM")
	CountryCodeDO                 = CountryCode("DO")
	CountryCodeEC                 = CountryCode("EC")
	CountryCodeEG                 = CountryCode("EG")
	CountryCodeSV                 = CountryCode("SV")
	CountryCodeGQ                 = CountryCode("GQ")
	CountryCodeER                 = CountryCode("ER")
	CountryCodeEE                 = CountryCode("EE")
	CountryCodeET                 = CountryCode("ET")
	CountryCodeFK                 = CountryCode("FK")
	CountryCodeFO                 = CountryCode("FO")
	CountryCodeFJ                 = CountryCode("FJ")
	CountryCodeFI                 = CountryCode("FI")
	CountryCodeFR                 = CountryCode("FR")
	CountryCodeGF                 = CountryCode("GF")
	CountryCodePF                 = CountryCode("PF")
	CountryCodeTF                 = CountryCode("TF")
	CountryCodeGA                 = CountryCode("GA")
	CountryCodeGM                 = CountryCode("GM")
	CountryCodeGE                 = CountryCode("GE")
	CountryCodeDE                 = CountryCode("DE")
	CountryCodeGH                 = CountryCode("GH")
	CountryCodeGI                 = CountryCode("GI")
	CountryCodeGR                 = CountryCode("GR")
	CountryCodeGL                 = CountryCode("GL")
	CountryCodeGD                 = CountryCode("GD")
	CountryCodeGP                 = CountryCode("GP")
	CountryCodeGU                 = CountryCode("GU")
	CountryCodeGT                 = CountryCode("GT")
	CountryCodeGG                 = CountryCode("GG")
	CountryCodeGN                 = CountryCode("GN")
	CountryCodeGW                 = CountryCode("GW")
	CountryCodeGY                 = CountryCode("GY")
	CountryCodeHT                 = CountryCode("HT")
	CountryCodeHM                 = CountryCode("HM")
	CountryCodeVA                 = CountryCode("VA")
	CountryCodeHN                 = CountryCode("HN")
	CountryCodeHK                 = CountryCode("HK")
	CountryCodeHU                 = CountryCode("HU")
	CountryCodeIS                 = CountryCode("IS")
	CountryCodeIN                 = CountryCode("IN")
	CountryCodeID                 = CountryCode("ID")
	CountryCodeIR                 = CountryCode("IR")
	CountryCodeIQ                 = CountryCode("IQ")
	CountryCodeIE                 = CountryCode("IE")
	CountryCodeIM                 = CountryCode("IM")
	CountryCodeIL                 = CountryCode("IL")
	CountryCodeIT                 = CountryCode("IT")
	CountryCodeJM                 = CountryCode("JM")
	CountryCodeJP                 = CountryCode("JP")
	CountryCodeJE                 = CountryCode("JE")
	CountryCodeJO                 = CountryCode("JO")
	CountryCodeKZ                 = CountryCode("KZ")
	CountryCodeKE                 = CountryCode("KE")
	CountryCodeKI                 = CountryCode("KI")
	CountryCodeKP                 = CountryCode("KP")
	CountryCodeKR                 = CountryCode("KR")
	CountryCodeKW                 = CountryCode("KW")
	CountryCodeKG                 = CountryCode("KG")
	CountryCodeLA                 = CountryCode("LA")
	CountryCodeLV                 = CountryCode("LV")
	CountryCodeLB                 = CountryCode("LB")
	CountryCodeLS                 = CountryCode("LS")
	CountryCodeLR                 = CountryCode("LR")
	CountryCodeLY                 = CountryCode("LY")
	CountryCodeLI                 = CountryCode("LI")
	CountryCodeLT                 = CountryCode("LT")
	CountryCodeLU                 = CountryCode("LU")
	CountryCodeMO                 = CountryCode("MO")
	CountryCodeMK                 = CountryCode("MK")
	CountryCodeMG                 = CountryCode("MG")
	CountryCodeMW                 = CountryCode("MW")
	CountryCodeMY                 = CountryCode("MY")
	CountryCodeMV                 = CountryCode("MV")
	CountryCodeML                 = CountryCode("ML")
	CountryCodeMT                 = CountryCode("MT")
	CountryCodeMH                 = CountryCode("MH")
	CountryCodeMQ                 = CountryCode("MQ")
	CountryCodeMR                 = CountryCode("MR")
	CountryCodeMU                 = CountryCode("MU")
	CountryCodeYT                 = CountryCode("YT")
	CountryCodeMX                 = CountryCode("MX")
	CountryCodeFM                 = CountryCode("FM")
	CountryCodeMD                 = CountryCode("MD")
	CountryCodeMC                 = CountryCode("MC")
	CountryCodeMN                 = CountryCode("MN")
	CountryCodeME                 = CountryCode("ME")
	CountryCodeMS                 = CountryCode("MS")
	CountryCodeMA                 = CountryCode("MA")
	CountryCodeMZ                 = CountryCode("MZ")
	CountryCodeMM                 = CountryCode("MM")
	CountryCodeNA                 = CountryCode("NA")
	CountryCodeNR                 = CountryCode("NR")
	CountryCodeNP                 = CountryCode("NP")
	CountryCodeNL                 = CountryCode("NL")
	CountryCodeNC                 = CountryCode("NC")
	CountryCodeNZ                 = CountryCode("NZ")
	CountryCodeNI                 = CountryCode("NI")
	CountryCodeNE                 = CountryCode("NE")
	CountryCodeNG                 = CountryCode("NG")
	CountryCodeNU                 = CountryCode("NU")
	CountryCodeNF                 = CountryCode("NF")
	CountryCodeMP                 = CountryCode("MP")
	CountryCodeNO                 = CountryCode("NO")
	CountryCodeOM                 = CountryCode("OM")
	CountryCodePK                 = CountryCode("PK")
	CountryCodePW                 = CountryCode("PW")
	CountryCodePS                 = CountryCode("PS")
	CountryCodePA                 = CountryCode("PA")
	CountryCodePG                 = CountryCode("PG")
	CountryCodePY                 = CountryCode("PY")
	CountryCodePE                 = CountryCode("PE")
	CountryCodePH                 = CountryCode("PH")
	CountryCodePN                 = CountryCode("PN")
	CountryCodePL                 = CountryCode("PL")
	CountryCodePT                 = CountryCode("PT")
	CountryCodePR                 = CountryCode("PR")
	CountryCodeQA                 = CountryCode("QA")
	CountryCodeRE                 = CountryCode("RE")
	CountryCodeRO                 = CountryCode("RO")
	CountryCodeRU                 = CountryCode("RU")
	CountryCodeRW                 = CountryCode("RW")
	CountryCodeBL                 = CountryCode("BL")
	CountryCodeSH                 = CountryCode("SH")
	CountryCodeKN                 = CountryCode("KN")
	CountryCodeLC                 = CountryCode("LC")
	CountryCodeMF                 = CountryCode("MF")
	CountryCodePM                 = CountryCode("PM")
	CountryCodeVC                 = CountryCode("VC")
	CountryCodeWS                 = CountryCode("WS")
	CountryCodeSM                 = CountryCode("SM")
	CountryCodeST                 = CountryCode("ST")
	CountryCodeSA                 = CountryCode("SA")
	CountryCodeSN                 = CountryCode("SN")
	CountryCodeRS                 = CountryCode("RS")
	CountryCodeSC                 = CountryCode("SC")
	CountryCodeSL                 = CountryCode("SL")
	CountryCodeSG                 = CountryCode("SG")
	CountryCodeSX                 = CountryCode("SX")
	CountryCodeSK                 = CountryCode("SK")
	CountryCodeSI                 = CountryCode("SI")
	CountryCodeSB                 = CountryCode("SB")
	CountryCodeSO                 = CountryCode("SO")
	CountryCodeZA                 = CountryCode("ZA")
	CountryCodeGS                 = CountryCode("GS")
	CountryCodeSS                 = CountryCode("SS")
	CountryCodeES                 = CountryCode("ES")
	CountryCodeLK                 = CountryCode("LK")
	CountryCodeSD                 = CountryCode("SD")
	CountryCodeSR                 = CountryCode("SR")
	CountryCodeSJ                 = CountryCode("SJ")
	CountryCodeSZ                 = CountryCode("SZ")
	CountryCodeSE                 = CountryCode("SE")
	CountryCodeCH                 = CountryCode("CH")
	CountryCodeSY                 = CountryCode("SY")
	CountryCodeTW                 = CountryCode("TW")
	CountryCodeTJ                 = CountryCode("TJ")
	CountryCodeTZ                 = CountryCode("TZ")
	CountryCodeTH                 = CountryCode("TH")
	CountryCodeTL                 = CountryCode("TL")
	CountryCodeTG                 = CountryCode("TG")
	CountryCodeTK                 = CountryCode("TK")
	CountryCodeTO                 = CountryCode("TO")
	CountryCodeTT                 = CountryCode("TT")
	CountryCodeTN                 = CountryCode("TN")
	CountryCodeTR                 = CountryCode("TR")
	CountryCodeTM                 = CountryCode("TM")
	CountryCodeTC                 = CountryCode("TC")
	CountryCodeTV                 = CountryCode("TV")
	CountryCodeUG                 = CountryCode("UG")
	CountryCodeUA                 = CountryCode("UA")
	CountryCodeAE                 = CountryCode("AE")
	CountryCodeGB                 = CountryCode("GB")
	CountryCodeUS                 = CountryCode("US")
	CountryCodeUM                 = CountryCode("UM")
	CountryCodeUY                 = CountryCode("UY")
	CountryCodeUZ                 = CountryCode("UZ")
	CountryCodeVU                 = CountryCode("VU")
	CountryCodeVE                 = CountryCode("VE")
	CountryCodeVN                 = CountryCode("VN")
	CountryCodeVG                 = CountryCode("VG")
	CountryCodeVI                 = CountryCode("VI")
	CountryCodeWF                 = CountryCode("WF")
	CountryCodeEH                 = CountryCode("EH")
	CountryCodeYE                 = CountryCode("YE")
	CountryCodeZM                 = CountryCode("ZM")
	CountryCodeZW                 = CountryCode("ZW")
)

func (enum CountryCode) String() string {
	if enum == "" {
		// return default value if empty
		return string(CountryCodeUnknownCountryCode)
	}
	return string(enum)
}

func (enum CountryCode) Values() []CountryCode {
	return []CountryCode{
		"unknown_country_code",
		"AF",
		"AX",
		"AL",
		"DZ",
		"AS",
		"AD",
		"AO",
		"AI",
		"AQ",
		"AG",
		"AR",
		"AM",
		"AW",
		"AU",
		"AT",
		"AZ",
		"BS",
		"BH",
		"BD",
		"BB",
		"BY",
		"BE",
		"BZ",
		"BJ",
		"BM",
		"BT",
		"BO",
		"BQ",
		"BA",
		"BW",
		"BV",
		"BR",
		"IO",
		"BN",
		"BG",
		"BF",
		"BI",
		"KH",
		"CM",
		"CA",
		"CV",
		"KY",
		"CF",
		"TD",
		"CL",
		"CN",
		"CX",
		"CC",
		"CO",
		"KM",
		"CG",
		"CD",
		"CK",
		"CR",
		"CI",
		"HR",
		"CU",
		"CW",
		"CY",
		"CZ",
		"DK",
		"DJ",
		"DM",
		"DO",
		"EC",
		"EG",
		"SV",
		"GQ",
		"ER",
		"EE",
		"ET",
		"FK",
		"FO",
		"FJ",
		"FI",
		"FR",
		"GF",
		"PF",
		"TF",
		"GA",
		"GM",
		"GE",
		"DE",
		"GH",
		"GI",
		"GR",
		"GL",
		"GD",
		"GP",
		"GU",
		"GT",
		"GG",
		"GN",
		"GW",
		"GY",
		"HT",
		"HM",
		"VA",
		"HN",
		"HK",
		"HU",
		"IS",
		"IN",
		"ID",
		"IR",
		"IQ",
		"IE",
		"IM",
		"IL",
		"IT",
		"JM",
		"JP",
		"JE",
		"JO",
		"KZ",
		"KE",
		"KI",
		"KP",
		"KR",
		"KW",
		"KG",
		"LA",
		"LV",
		"LB",
		"LS",
		"LR",
		"LY",
		"LI",
		"LT",
		"LU",
		"MO",
		"MK",
		"MG",
		"MW",
		"MY",
		"MV",
		"ML",
		"MT",
		"MH",
		"MQ",
		"MR",
		"MU",
		"YT",
		"MX",
		"FM",
		"MD",
		"MC",
		"MN",
		"ME",
		"MS",
		"MA",
		"MZ",
		"MM",
		"NA",
		"NR",
		"NP",
		"NL",
		"NC",
		"NZ",
		"NI",
		"NE",
		"NG",
		"NU",
		"NF",
		"MP",
		"NO",
		"OM",
		"PK",
		"PW",
		"PS",
		"PA",
		"PG",
		"PY",
		"PE",
		"PH",
		"PN",
		"PL",
		"PT",
		"PR",
		"QA",
		"RE",
		"RO",
		"RU",
		"RW",
		"BL",
		"SH",
		"KN",
		"LC",
		"MF",
		"PM",
		"VC",
		"WS",
		"SM",
		"ST",
		"SA",
		"SN",
		"RS",
		"SC",
		"SL",
		"SG",
		"SX",
		"SK",
		"SI",
		"SB",
		"SO",
		"ZA",
		"GS",
		"SS",
		"ES",
		"LK",
		"SD",
		"SR",
		"SJ",
		"SZ",
		"SE",
		"CH",
		"SY",
		"TW",
		"TJ",
		"TZ",
		"TH",
		"TL",
		"TG",
		"TK",
		"TO",
		"TT",
		"TN",
		"TR",
		"TM",
		"TC",
		"TV",
		"UG",
		"UA",
		"AE",
		"GB",
		"US",
		"UM",
		"UY",
		"UZ",
		"VU",
		"VE",
		"VN",
		"VG",
		"VI",
		"WF",
		"EH",
		"YE",
		"ZM",
		"ZW",
	}
}

func (enum CountryCode) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *CountryCode) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = CountryCode(CountryCode(tmp).String())
	return nil
}

type LanguageCode string

const (
	LanguageCodeUnknownLanguageCode = LanguageCode("unknown_language_code")
	LanguageCodeEnUS                = LanguageCode("en_US")
	LanguageCodeFrFR                = LanguageCode("fr_FR")
	LanguageCodeDeDE                = LanguageCode("de_DE")
)

func (enum LanguageCode) String() string {
	if enum == "" {
		// return default value if empty
		return string(LanguageCodeUnknownLanguageCode)
	}
	return string(enum)
}

func (enum LanguageCode) Values() []LanguageCode {
	return []LanguageCode{
		"unknown_language_code",
		"en_US",
		"fr_FR",
		"de_DE",
	}
}

func (enum LanguageCode) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *LanguageCode) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = LanguageCode(LanguageCode(tmp).String())
	return nil
}
