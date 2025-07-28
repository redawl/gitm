package internal

// FilterToken is a token parsed from the packet filter string
//
// It is used to determining which packets should be displayed
// in the ui for a given filter.
// Ex: "status:101" would match any http packets with a response status of "101"
type FilterToken struct {
	// FilterType is the type of filter.
	// This is the part before the : in a FilterToken
	FilterType string
	// Negate is whether the FilterToken causes the inverse of its meaning.
	// This is set by having a "-" before FilterContent.
	//
	// Ex: "status:-101"
	Negate bool
	// FilterContent is the content of the FilterToken.
	// This is the part after the : in a FilterToken
	FilterContent string
}
