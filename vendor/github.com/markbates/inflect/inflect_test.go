package inflect

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// test data

var SingularToPlural = map[string]string{
	"search":      "searches",
	"switch":      "switches",
	"fix":         "fixes",
	"box":         "boxes",
	"process":     "processes",
	"address":     "addresses",
	"case":        "cases",
	"stack":       "stacks",
	"wish":        "wishes",
	"fish":        "fish",
	"jeans":       "jeans",
	"funky jeans": "funky jeans",
	"category":    "categories",
	"query":       "queries",
	"ability":     "abilities",
	"agency":      "agencies",
	"movie":       "movies",
	"archive":     "archives",
	"index":       "indices",
	"wife":        "wives",
	"safe":        "saves",
	"half":        "halves",
	"move":        "moves",
	"salesperson": "salespeople",
	"person":      "people",
	"spokesman":   "spokesmen",
	"man":         "men",
	"woman":       "women",
	"basis":       "bases",
	"diagnosis":   "diagnoses",
	"diagnosis_a": "diagnosis_as",
	"datum":       "data",
	"medium":      "media",
	"stadium":     "stadia",
	"analysis":    "analyses",
	"node_child":  "node_children",
	"child":       "children",
	"experience":  "experiences",
	"day":         "days",
	"comment":     "comments",
	"foobar":      "foobars",
	"newsletter":  "newsletters",
	"old_news":    "old_news",
	"news":        "news",
	"series":      "series",
	"species":     "species",
	"quiz":        "quizzes",
	"perspective": "perspectives",
	"ox":          "oxen",
	"photo":       "photos",
	"buffalo":     "buffaloes",
	"tomato":      "tomatoes",
	"dwarf":       "dwarves",
	"elf":         "elves",
	"information": "information",
	"equipment":   "equipment",
	"bus":         "buses",
	"status":      "statuses",
	"Status":      "Statuses",
	"status_code": "status_codes",
	"mouse":       "mice",
	"louse":       "lice",
	"house":       "houses",
	"octopus":     "octopi",
	"virus":       "viri",
	"alias":       "aliases",
	"portfolio":   "portfolios",
	"vertex":      "vertices",
	"matrix":      "matrices",
	"matrix_fu":   "matrix_fus",
	"axis":        "axes",
	"testis":      "testes",
	"crisis":      "crises",
	"rice":        "rice",
	"shoe":        "shoes",
	"horse":       "horses",
	"prize":       "prizes",
	"edge":        "edges",
	"database":    "databases",
}

var CapitalizeMixture = map[string]string{
	"product":               "Product",
	"special_guest":         "Special_guest",
	"applicationController": "ApplicationController",
	"Area51Controller":      "Area51Controller",
	"id":                    "ID",
	"SQL":                   "SQL",
	"sql":                   "SQL",
	"sQL":                   "SQL",
}

var CamelToUnderscore = map[string]string{
	"Product":               "product",
	"SpecialGuest":          "special_guest",
	"ApplicationController": "application_controller",
	"Area51Controller":      "area51_controller",
}

var UnderscoreToLowerCamel = map[string]string{
	"product":                "product",
	"special_guest":          "specialGuest",
	"application_controller": "applicationController",
	"area51_controller":      "area51Controller",
}

var CamelToUnderscoreWithoutReverse = map[string]string{
	"HTMLTidy":          "html_tidy",
	"HTMLTidyGenerator": "html_tidy_generator",
	"FreeBsd":           "free_bsd",
	"HTML":              "html",
}

var ClassNameToForeignKeyWithUnderscore = map[string]string{
	"Person":  "person_id",
	"Account": "account_id",
}

var PluralToForeignKeyWithUnderscore = map[string]string{
	"people":   "person_id",
	"accounts": "account_id",
}

var ClassNameToForeignKeyWithoutUnderscore = map[string]string{
	"Person":  "personid",
	"Account": "accountid",
}

var ClassNameToTableName = map[string]string{
	"PrimarySpokesman": "primary_spokesmen",
	"NodeChild":        "node_children",
	"Alias":            "aliases",
}

var StringToParameterized = map[string]string{
	"Donald E. Knuth":                     "donald-e-knuth",
	"Random text with *(bad)* characters": "random-text-with-bad-characters",
	"Allow_Under_Scores":                  "allow_under_scores",
	"Trailing bad characters!@#":          "trailing-bad-characters",
	"!@#Leading bad characters":           "leading-bad-characters",
	"Squeeze   separators":                "squeeze-separators",
	"Test with + sign":                    "test-with-sign",
	"Test with malformed utf8 \251":       "test-with-malformed-utf8",
}

var StringToParameterizeWithNoSeparator = map[string]string{
	"Donald E. Knuth":                     "donaldeknuth",
	"With-some-dashes":                    "with-some-dashes",
	"Random text with *(bad)* characters": "randomtextwithbadcharacters",
	"Trailing bad characters!@#":          "trailingbadcharacters",
	"!@#Leading bad characters":           "leadingbadcharacters",
	"Squeeze   separators":                "squeezeseparators",
	"Test with + sign":                    "testwithsign",
	"Test with malformed utf8 \251":       "testwithmalformedutf8",
}

var StringToParameterizeWithUnderscore = map[string]string{
	"Donald E. Knuth":                     "donald_e_knuth",
	"Random text with *(bad)* characters": "random_text_with_bad_characters",
	"With-some-dashes":                    "with-some-dashes",
	"Retain_underscore":                   "retain_underscore",
	"Trailing bad characters!@#":          "trailing_bad_characters",
	"!@#Leading bad characters":           "leading_bad_characters",
	"Squeeze   separators":                "squeeze_separators",
	"Test with + sign":                    "test_with_sign",
	"Test with malformed utf8 \251":       "test_with_malformed_utf8",
}

var StringToParameterizedAndNormalized = map[string]string{
	"Malmö":         "malmo",
	"Garçons":       "garcons",
	"Opsů":          "opsu",
	"Ærøskøbing":    "aeroskobing",
	"Aßlar":         "asslar",
	"Japanese: 日本語": "japanese",
}

var UnderscoreToHuman = map[string]string{
	"employee_salary": "Employee salary",
	"employee_id":     "Employee",
	"underground":     "Underground",
	"óbito":           "Óbito",
}

var MixtureToTitleCase = map[string]string{
	"active_record":       "Active Record",
	"ActiveRecord":        "Active Record",
	"action web service":  "Action Web Service",
	"Action Web Service":  "Action Web Service",
	"Action web service":  "Action Web Service",
	"actionwebservice":    "Actionwebservice",
	"Actionwebservice":    "Actionwebservice",
	"david's code":        "David's Code",
	"David's code":        "David's Code",
	"david's Code":        "David's Code",
	"my_cool_URL_enabled": "My Cool URL Enabled",
	"service_API_URL":     "Service API URL",
}

var OrdinalNumbers = map[string]string{
	"-1":    "-1st",
	"-2":    "-2nd",
	"-3":    "-3rd",
	"-4":    "-4th",
	"-5":    "-5th",
	"-6":    "-6th",
	"-7":    "-7th",
	"-8":    "-8th",
	"-9":    "-9th",
	"-10":   "-10th",
	"-11":   "-11th",
	"-12":   "-12th",
	"-13":   "-13th",
	"-14":   "-14th",
	"-20":   "-20th",
	"-21":   "-21st",
	"-22":   "-22nd",
	"-23":   "-23rd",
	"-24":   "-24th",
	"-100":  "-100th",
	"-101":  "-101st",
	"-102":  "-102nd",
	"-103":  "-103rd",
	"-104":  "-104th",
	"-110":  "-110th",
	"-111":  "-111th",
	"-112":  "-112th",
	"-113":  "-113th",
	"-1000": "-1000th",
	"-1001": "-1001st",
	"0":     "0th",
	"1":     "1st",
	"2":     "2nd",
	"3":     "3rd",
	"4":     "4th",
	"5":     "5th",
	"6":     "6th",
	"7":     "7th",
	"8":     "8th",
	"9":     "9th",
	"10":    "10th",
	"11":    "11th",
	"12":    "12th",
	"13":    "13th",
	"14":    "14th",
	"20":    "20th",
	"21":    "21st",
	"22":    "22nd",
	"23":    "23rd",
	"24":    "24th",
	"100":   "100th",
	"101":   "101st",
	"102":   "102nd",
	"103":   "103rd",
	"104":   "104th",
	"110":   "110th",
	"111":   "111th",
	"112":   "112th",
	"113":   "113th",
	"1000":  "1000th",
	"1001":  "1001st",
}

var UnderscoresToDashes = map[string]string{
	"street":                "street",
	"street_address":        "street-address",
	"person_street_address": "person-street-address",
}

var Irregularities = map[string]string{
	"person": "people",
	"man":    "men",
	"child":  "children",
	"sex":    "sexes",
	"move":   "moves",
}

type AcronymCase struct {
	camel string
	under string
	human string
	title string
}

var AcronymCases = []*AcronymCase{
	//           camelize             underscore            humanize              titleize
	&AcronymCase{"API", "api", "API", "API"},
	&AcronymCase{"APIController", "api_controller", "API controller", "API Controller"},
	&AcronymCase{"Nokogiri::HTML", "nokogiri/html", "Nokogiri/HTML", "Nokogiri/HTML"},
	&AcronymCase{"HTTPAPI", "http_api", "HTTP API", "HTTP API"},
	&AcronymCase{"HTTP::Get", "http/get", "HTTP/get", "HTTP/Get"},
	&AcronymCase{"SSLError", "ssl_error", "SSL error", "SSL Error"},
	&AcronymCase{"RESTful", "restful", "RESTful", "RESTful"},
	&AcronymCase{"RESTfulController", "restful_controller", "RESTful controller", "RESTful Controller"},
	&AcronymCase{"IHeartW3C", "i_heart_w3c", "I heart W3C", "I Heart W3C"},
	&AcronymCase{"PhDRequired", "phd_required", "PhD required", "PhD Required"},
	&AcronymCase{"IRoRU", "i_ror_u", "I RoR u", "I RoR U"},
	&AcronymCase{"RESTfulHTTPAPI", "restful_http_api", "RESTful HTTP API", "RESTful HTTP API"},
	// misdirection
	&AcronymCase{"Capistrano", "capistrano", "Capistrano", "Capistrano"},
	&AcronymCase{"CapiController", "capi_controller", "Capi controller", "Capi Controller"},
	&AcronymCase{"HttpsApis", "https_apis", "Https apis", "Https Apis"},
	&AcronymCase{"Html5", "html5", "Html5", "Html5"},
	&AcronymCase{"Restfully", "restfully", "Restfully", "Restfully"},
	&AcronymCase{"RoRails", "ro_rails", "Ro rails", "Ro Rails"},
}

// tests

func Test_LoadViaFile(t *testing.T) {
	require.Equal(t, "feedback", Pluralize("feedback"))
	require.Equal(t, "buffalo!", Singularize("buffalos!"))
}

func TestForeignKeyToAttribute(t *testing.T) {
	require.Equal(t, "PersonID", ForeignKeyToAttribute("person_id"))
	require.Equal(t, "ID", ForeignKeyToAttribute("id"))
}

func TestPluralizeWithSize(t *testing.T) {
	require.Equal(t, "plurals", PluralizeWithSize("plurals", 2))
	require.Equal(t, "plurals", PluralizeWithSize("plurals", 0))
	require.Equal(t, "plural", PluralizeWithSize("plurals", 1))
}

func TestPluralizePlurals(t *testing.T) {
	require.Equal(t, "plurals", Pluralize("plurals"))
	require.Equal(t, "Plurals", Pluralize("Plurals"))
}

func TestPluralizeEmptyString(t *testing.T) {
	require.Equal(t, "", Pluralize(""))
}

func TestUncountables(t *testing.T) {
	for word := range Uncountables() {
		require.Equal(t, word, Singularize(word))
		require.Equal(t, word, Pluralize(word))
		require.Equal(t, Pluralize(word), Singularize(word))
	}
}

func TestUncountableWordIsNotGreedy(t *testing.T) {
	uncountableWord := "ors"
	countableWord := "sponsor"

	AddUncountable(uncountableWord)

	require.Equal(t, uncountableWord, Singularize(uncountableWord))
	require.Equal(t, uncountableWord, Pluralize(uncountableWord))
	require.Equal(t, Pluralize(uncountableWord), Singularize(uncountableWord))
	require.Equal(t, "sponsor", Singularize(countableWord))
	require.Equal(t, "sponsors", Pluralize(countableWord))
	require.Equal(t, "sponsor", Singularize(Pluralize(countableWord)))
}

func TestPluralizeSingular(t *testing.T) {
	for singular, plural := range SingularToPlural {
		require.Equal(t, plural, Pluralize(singular))
		require.Equal(t, Capitalize(plural), Capitalize(Pluralize(singular)))
	}
}

func TestSingularizePlural(t *testing.T) {
	for singular, plural := range SingularToPlural {
		require.Equal(t, singular, Singularize(plural))
		require.Equal(t, Capitalize(singular), Capitalize(Singularize(plural)))
	}
}

func TestSingularizeSingular(t *testing.T) {
	for singular := range SingularToPlural {
		require.Equal(t, singular, Singularize(singular))
		require.Equal(t, Capitalize(singular), Capitalize(Singularize(singular)))
	}
}

func TestPluralizePlural(t *testing.T) {
	for _, plural := range SingularToPlural {
		require.Equal(t, plural, Pluralize(plural))
		require.Equal(t, Capitalize(plural), Capitalize(Pluralize(plural)))
	}
}

func TestOverwritePreviousInflectors(t *testing.T) {
	require.Equal(t, "series", Singularize("series"))
	AddSingular("series", "serie")
	require.Equal(t, "serie", Singularize("series"))
	AddUncountable("series") // reset
}

func TestTitleize(t *testing.T) {
	for before, titleized := range MixtureToTitleCase {
		require.Equal(t, titleized, Titleize(before))
	}
}

func TestCapitalize(t *testing.T) {
	for lower, capitalized := range CapitalizeMixture {
		require.Equal(t, capitalized, Capitalize(lower))
	}
}

func TestCamelize(t *testing.T) {
	for camel, underscore := range CamelToUnderscore {
		require.Equal(t, camel, Camelize(underscore))
	}
}

func TestCamelizeWithLowerDowncasesTheFirstLetter(t *testing.T) {
	require.Equal(t, "capital", CamelizeDownFirst("Capital"))
}

func TestCamelizeWithUnderscores(t *testing.T) {
	require.Equal(t, "CamelCase", Camelize("Camel_Case"))
}

// func TestAcronyms(t *testing.T) {
//     AddAcronym("API")
//     AddAcronym("HTML")
//     AddAcronym("HTTP")
//     AddAcronym("RESTful")
//     AddAcronym("W3C")
//     AddAcronym("PhD")
//     AddAcronym("RoR")
//     AddAcronym("SSL")
//     // each in table
//     for _,x := range AcronymCases {
//         require.Equal(t, x.camel, Camelize(x.under))
//         require.Equal(t, x.camel, Camelize(x.camel))
//         require.Equal(t, x.under, Underscore(x.under))
//         require.Equal(t, x.under, Underscore(x.camel))
//         require.Equal(t, x.title, Titleize(x.under))
//         require.Equal(t, x.title, Titleize(x.camel))
//         require.Equal(t, x.human, Humanize(x.under))
//     }
// }

// func TestAcronymOverride(t *testing.T) {
//     AddAcronym("API")
//     AddAcronym("LegacyApi")
//     require.Equal(t, "LegacyApi", Camelize("legacyapi"))
//     require.Equal(t, "LegacyAPI", Camelize("legacy_api"))
//     require.Equal(t, "SomeLegacyApi", Camelize("some_legacyapi"))
//     require.Equal(t, "Nonlegacyapi", Camelize("nonlegacyapi"))
// }

// func TestAcronymsCamelizeLower(t *testing.T) {
//     AddAcronym("API")
//     AddAcronym("HTML")
//     require.Equal(t, "htmlAPI", CamelizeDownFirst("html_api"))
//     require.Equal(t, "htmlAPI", CamelizeDownFirst("htmlAPI"))
//     require.Equal(t, "htmlAPI", CamelizeDownFirst("HTMLAPI"))
// }

func TestUnderscoreAcronymSequence(t *testing.T) {
	AddAcronym("API")
	AddAcronym("HTML5")
	AddAcronym("HTML")
	require.Equal(t, "html5_html_api", Underscore("HTML5HTMLAPI"))
}

func TestUnderscore(t *testing.T) {
	for camel, underscore := range CamelToUnderscore {
		require.Equal(t, underscore, Underscore(camel))
	}
	for camel, underscore := range CamelToUnderscoreWithoutReverse {
		require.Equal(t, underscore, Underscore(camel))
	}
}

func TestForeignKey(t *testing.T) {
	for klass, foreignKey := range ClassNameToForeignKeyWithUnderscore {
		require.Equal(t, foreignKey, ForeignKey(klass))
	}
	for word, foreignKey := range PluralToForeignKeyWithUnderscore {
		require.Equal(t, foreignKey, ForeignKey(word))
	}
	for klass, foreignKey := range ClassNameToForeignKeyWithoutUnderscore {
		require.Equal(t, foreignKey, ForeignKeyCondensed(klass))
	}
}

func TestTableize(t *testing.T) {
	for klass, table := range ClassNameToTableName {
		require.Equal(t, table, Tableize(klass))
	}
}

func TestParameterize(t *testing.T) {
	for str, parameterized := range StringToParameterized {
		require.Equal(t, parameterized, Parameterize(str))
	}
}

func TestParameterizeAndNormalize(t *testing.T) {
	for str, parameterized := range StringToParameterizedAndNormalized {
		require.Equal(t, parameterized, Parameterize(str))
	}
}

func TestParameterizeWithCustomSeparator(t *testing.T) {
	for str, parameterized := range StringToParameterizeWithUnderscore {
		require.Equal(t, parameterized, ParameterizeJoin(str, "_"))
	}
}

func TestTypeify(t *testing.T) {
	for klass, table := range ClassNameToTableName {
		require.Equal(t, klass, Typeify(table))
		require.Equal(t, klass, Typeify("table_prefix."+table))
	}
}

func TestTypeifyWithLeadingSchemaName(t *testing.T) {
	require.Equal(t, "FooBar", Typeify("schema.foo_bar"))
}

func TestHumanize(t *testing.T) {
	for underscore, human := range UnderscoreToHuman {
		require.Equal(t, human, Humanize(underscore))
	}
}

func TestHumanizeByString(t *testing.T) {
	AddHuman("col_rpted_bugs", "reported bugs")
	require.Equal(t, "90 reported bugs recently", Humanize("90 col_rpted_bugs recently"))
}

func TestOrdinal(t *testing.T) {
	for number, ordinalized := range OrdinalNumbers {
		require.Equal(t, ordinalized, Ordinalize(number))
	}
}

func TestDasherize(t *testing.T) {
	for underscored, dasherized := range UnderscoresToDashes {
		require.Equal(t, dasherized, Dasherize(underscored))
	}
}

func TestUnderscoreAsReverseOfDasherize(t *testing.T) {
	for underscored := range UnderscoresToDashes {
		require.Equal(t, underscored, Underscore(Dasherize(underscored)))
	}
}

func TestUnderscoreToLowerCamel(t *testing.T) {
	for underscored, lower := range UnderscoreToLowerCamel {
		require.Equal(t, lower, CamelizeDownFirst(underscored))
	}
}

func Test_clear_all(t *testing.T) {
	// test a way of resetting inflexions
}

func TestIrregularityBetweenSingularAndPlural(t *testing.T) {
	for singular, plural := range Irregularities {
		AddIrregular(singular, plural)
		require.Equal(t, singular, Singularize(plural))
		require.Equal(t, plural, Pluralize(singular))
	}
}

func TestPluralizeOfIrregularity(t *testing.T) {
	for singular, plural := range Irregularities {
		AddIrregular(singular, plural)
		require.Equal(t, plural, Pluralize(plural))
	}
}

func Test_Address(t *testing.T) {
	require.Equal(t, "address", Singularize("address"))
	require.Equal(t, "addresses", Pluralize("address"))
	require.Equal(t, "address", Singularize("addresses"))
	require.Equal(t, "addresses", Pluralize("addresses"))
}
