package ui

import (
	"testing"
)

func TestGetTokens(t *testing.T) {
	filterString := "host:google.com statuscode:200"

	filterPairs := getTokens(filterString)

	if len(filterPairs) != 2 {
		t.Errorf("len(getTokens(\"%s\") = %d, expected 2", filterString, len(filterPairs))
	} else if filterPairs[0].filterType != "host" || filterPairs[0].filterContent != "google.com" {
		t.Errorf("filterType = \"%s\" && filterContent = \"%s\", expected host && google.com", filterPairs[0].filterType, filterPairs[0].filterContent)
	} else if filterPairs[1].filterType != "statuscode" || filterPairs[1].filterContent != "200" {
		t.Errorf("filterType = \"%s\" && filterContent = \"%s\", expected statuscode && 200", filterPairs[1].filterType, filterPairs[1].filterContent)
	}
}

func TestGetTokensNoColon(t *testing.T) {
	filterString := "whatfieldamifiltering"

	filterPairs := getTokens(filterString)

	if len(filterPairs) != 0 {
		t.Errorf("len(getTokens(\"%s\")) = %d, expected 0", filterString, len(filterPairs))
	}
}

func TestGetTokensInvalidPart(t *testing.T) {
	filterString := "host:google.com whatfieldamifiltering"

	filterPairs := getTokens(filterString)

	if len(filterPairs) != 1 {
		t.Errorf("len(getTokens(\"%s\")) = %d, expected 1", filterString, len(filterPairs))
		return
	}

	if filterPairs[0].filterType != "host" || filterPairs[0].filterContent != "google.com" {
		t.Errorf("filterType = \"%s\" && filterContent = \"%s\", expected host && google.com", filterPairs[0].filterType, filterPairs[0].filterContent)
	}
}

func TestGetTokensInvalidPart2(t *testing.T) {
	filterString := "whatfieldamifiltering host:google.com"

	filterPairs := getTokens(filterString)

	if len(filterPairs) != 1 {
		t.Errorf("len(getTokens(\"%s\")) = %d, expected 1", filterString, len(filterPairs))
		return
	}

	if filterPairs[0].filterType != "host" || filterPairs[0].filterContent != "google.com" {
		t.Errorf("filterType = \"%s\" && filterContent = \"%s\", expected host && google.com", filterPairs[0].filterType, filterPairs[0].filterContent)
	}
}

func TestGetTokensWithQuotes(t *testing.T) {
	filterString := "content:\"bob joe was here\""

	filterPairs := getTokens(filterString)

	if len(filterPairs) != 1 {
		t.Errorf("len(getTokens(\"%s\")) = %d, expected 1", filterString, len(filterPairs))
	} else if filterPairs[0].filterType != "content" || filterPairs[0].filterContent != "bob joe was here" {
		t.Errorf("filterType = \"%s\" && filterContent = \"%s\", expected content && \"bob joe was here\"", filterPairs[0].filterType, filterPairs[0].filterContent)
	}
}

func TestGetTokensOnlyFilterType(t *testing.T) {
	filterString := "host:"

	filterPairs := getTokens(filterString)

	if len(filterPairs) != 1 {
		t.Errorf("len(getTokens(\"%s\")) = %d. expected 1", filterString, len(filterPairs))
	} else if filterPairs[0].filterType != "host" || filterPairs[0].filterContent != "" {
		t.Errorf("filterType = \"%s\" && filterContent = \"%s\", expected host && \"\"", filterPairs[0].filterType, filterPairs[0].filterContent)
	}
}

func TestGetTokensNegate(t *testing.T) {
	filterString := "host:-google.com"

	filterPairs := getTokens(filterString)

	if len(filterPairs) != 1 {
		t.Errorf("len(getTokens(\"%s\")) = %d. expected 1", filterString, len(filterPairs))
	} else if filterPairs[0].filterType != "host" || filterPairs[0].filterContent != "google.com" || !filterPairs[0].negate {
		t.Errorf("filterType = \"%s\" && filterContent = \"%s\" && negate = %t, expected host && google.com && true", filterPairs[0].filterType, filterPairs[0].filterContent, filterPairs[0].negate)
	}
}

func TestGetTokensNegateEmpty(t *testing.T) {
	filterString := "host:-"

	filterPairs := getTokens(filterString)

	if len(filterPairs) != 1 {
		t.Errorf("len(getTokens(\"%s\")) = %d. expected 1", filterString, len(filterPairs))
	} else if filterPairs[0].filterType != "host" || filterPairs[0].filterContent != "" || !filterPairs[0].negate {
		t.Errorf("filterType = \"%s\" && filterContent = \"%s\" && negate = %t, expected host && \"\" && true", filterPairs[0].filterType, filterPairs[0].filterContent, filterPairs[0].negate)
	}
}
