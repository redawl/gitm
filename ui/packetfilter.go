package ui

import (
	"log/slog"
	"strings"

	"com.github.redawl.gitm/packet"
)

type filterPair struct {
    filterType string
    negate bool
    filterContent string
}

func getTokens (filterString string) []filterPair {
    filterStringStripped := strings.Trim(filterString, " ")

    filterPairs := make([]filterPair, 0)

    i := 0
    length := len(filterStringStripped)

    for i < length {
        if filterStringStripped[i] == ' ' {
            i++
            continue
        }

        fp := filterPair{}
        colonIndex := strings.Index(filterStringStripped[i:], ":")

        if colonIndex == -1 {
            return filterPairs
        } else {
            colonIndex += i
        }

        // Get filter type
        fp.filterType = filterStringStripped[i:colonIndex]
        if strings.Contains(fp.filterType, " ") {
            // Get rid of prev cruft
            spaceIndex := strings.Index(fp.filterType, " ")
            fp.filterType = fp.filterType[spaceIndex+1:]
        }

        if len(filterStringStripped) <= colonIndex + 1 {
            // found filterType without filterContent
            fp.negate = false
            filterPairs = append(filterPairs, fp)
            return filterPairs
        }
        
        // Get filter content
        if filterStringStripped[colonIndex + 1] == '-' {
            fp.negate = true
            colonIndex++
        } else {
            fp.negate = false
        }

        if len(filterStringStripped) <= colonIndex + 1 {
            // found filterType without filterContent
            fp.negate = false
            filterPairs = append(filterPairs, fp)
            return filterPairs
        }

        if filterStringStripped[colonIndex + 1] == '"' {
            quoteIndex := strings.Index(filterStringStripped[colonIndex+2:], "\"")
            if quoteIndex == -1 {
                spaceIndex := strings.Index(filterStringStripped[colonIndex+2:], " ")
                
                if spaceIndex == -1 {
                    fp.filterContent = filterStringStripped[colonIndex + 1:]
                    i = length
                } else {
                    fp.filterContent = filterStringStripped[colonIndex + 1:colonIndex + spaceIndex]
                    i = spaceIndex + colonIndex
                }
            } else {
                fp.filterContent = filterStringStripped[colonIndex + 2: colonIndex + quoteIndex + 2]
                i = quoteIndex + colonIndex + 1
            }
        } else {
            spaceIndex := strings.Index(filterStringStripped[colonIndex:], " ")
            
            if spaceIndex == -1 {
                fp.filterContent = filterStringStripped[colonIndex + 1:]
                i = length
            } else {
                fp.filterContent = filterStringStripped[colonIndex + 1:colonIndex + spaceIndex]
                i = spaceIndex + colonIndex
            }
        }

        filterPairs = append(filterPairs, fp)
    }

    return filterPairs
}

func FilterPackets (filterString string, packets []*packet.HttpPacket) []*packet.HttpPacket {
    filterPairs := getTokens(filterString)
    passedPackets := make([]*packet.HttpPacket, 0)

    for _, p := range packets {
        passed := true
        for _, filterPair := range filterPairs {
            switch filterPair.filterType {
                case "host": {
                    if filterPair.negate == (strings.Contains(p.ServerIp, filterPair.filterContent)) {
                        passed = false
                        break
                    }
                }
                case "statuscode": {
                    if filterPair.negate == (strings.Contains(p.Status, filterPair.filterContent)) {
                        passed = false
                        break
                    }
                }
                case "content": {
                    if filterPair.negate == (strings.Contains(string(p.RespContent), filterPair.filterContent)) {
                        passed = false
                        break
                    }
                }
                case "method": {
                    if filterPair.negate == (strings.Contains(p.Method, filterPair.filterContent)) {
                        passed = false
                        break
                    }
                }
                default: {
                    slog.Warn("Unknown filter specified", "filterType", filterPair.filterType, "filterContent", filterPair.filterContent)
                }
            }
        }
        // Passed all filters
        if passed {
            passedPackets = append(passedPackets, p)
        } else {
            passed = true
        }
    }

    return passedPackets
}
