package ui

import (
	"testing"
)

func TestGetTokens(t *testing.T) {
	filterString := "host:google.com statuscode:200"

	tokens := getTokens(filterString)

	if len(tokens) != 2 {
		t.Errorf("len(getTokens(\"%s\") = %d, expected 2", filterString, len(tokens))
	} else if tokens[0].FilterType != "host" || tokens[0].FilterContent != "google.com" {
		t.Errorf("filterType = \"%s\" && filterContent = \"%s\", expected host && google.com", tokens[0].FilterType, tokens[0].FilterContent)
	} else if tokens[1].FilterType != "statuscode" || tokens[1].FilterContent != "200" {
		t.Errorf("filterType = \"%s\" && filterContent = \"%s\", expected statuscode && 200", tokens[1].FilterType, tokens[1].FilterContent)
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

	tokens := getTokens(filterString)

	if len(tokens) != 1 {
		t.Errorf("len(getTokens(\"%s\")) = %d, expected 1", filterString, len(tokens))
		return
	}

	if tokens[0].FilterType != "host" || tokens[0].FilterContent != "google.com" {
		t.Errorf("filterType = \"%s\" && filterContent = \"%s\", expected host && google.com", tokens[0].FilterType, tokens[0].FilterContent)
	}
}

func TestGetTokensInvalidPart2(t *testing.T) {
	filterString := "whatfieldamifiltering host:google.com"

	filterPairs := getTokens(filterString)

	if len(filterPairs) != 1 {
		t.Errorf("len(getTokens(\"%s\")) = %d, expected 1", filterString, len(filterPairs))
		return
	}

	if filterPairs[0].FilterType != "host" || filterPairs[0].FilterContent != "google.com" {
		t.Errorf("filterType = \"%s\" && filterContent = \"%s\", expected host && google.com", filterPairs[0].FilterType, filterPairs[0].FilterContent)
	}
}

func TestGetTokensWithQuotes(t *testing.T) {
	filterString := "content:\"bob joe was here\""

	filterPairs := getTokens(filterString)

	if len(filterPairs) != 1 {
		t.Errorf("len(getTokens(\"%s\")) = %d, expected 1", filterString, len(filterPairs))
	} else if filterPairs[0].FilterType != "content" || filterPairs[0].FilterContent != "bob joe was here" {
		t.Errorf("filterType = \"%s\" && filterContent = \"%s\", expected content && \"bob joe was here\"", filterPairs[0].FilterType, filterPairs[0].FilterContent)
	}
}

func TestGetTokensOnlyFilterType(t *testing.T) {
	filterString := "host:"

	filterPairs := getTokens(filterString)

	if len(filterPairs) != 1 {
		t.Errorf("len(getTokens(\"%s\")) = %d. expected 1", filterString, len(filterPairs))
	} else if filterPairs[0].FilterType != "host" || filterPairs[0].FilterContent != "" {
		t.Errorf("filterType = \"%s\" && filterContent = \"%s\", expected host && \"\"", filterPairs[0].FilterType, filterPairs[0].FilterContent)
	}
}

func TestGetTokensNegate(t *testing.T) {
	filterString := "host:-google.com"

	filterPairs := getTokens(filterString)

	if len(filterPairs) != 1 {
		t.Errorf("len(getTokens(\"%s\")) = %d. expected 1", filterString, len(filterPairs))
	} else if filterPairs[0].FilterType != "host" || filterPairs[0].FilterContent != "google.com" || !filterPairs[0].Negate {
		t.Errorf("filterType = \"%s\" && filterContent = \"%s\" && negate = %t, expected host && google.com && true", filterPairs[0].FilterType, filterPairs[0].FilterContent, filterPairs[0].Negate)
	}
}

func TestGetTokensNegateEmpty(t *testing.T) {
	filterString := "host:-"

	filterPairs := getTokens(filterString)

	if len(filterPairs) != 1 {
		t.Errorf("len(getTokens(\"%s\")) = %d. expected 1", filterString, len(filterPairs))
	} else if filterPairs[0].FilterType != "host" || filterPairs[0].FilterContent != "" || !filterPairs[0].Negate {
		t.Errorf("filterType = \"%s\" && filterContent = \"%s\" && negate = %t, expected host && \"\" && true", filterPairs[0].FilterType, filterPairs[0].FilterContent, filterPairs[0].Negate)
	}
}
