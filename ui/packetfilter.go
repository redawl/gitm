package ui

import (
	"log/slog"
	"strings"

	"com.github.redawl.gitm/packet"
)

const (
    FILTER_HOSTNAME  = "hostname"
    FILTER_METHOD    = "method"
    FILTER_PATH      = "path"
    FILTER_REQ_BODY  = "reqbody"
    // TODO filter on version?
    FILTER_STATUS    = "status"
    FILTER_RESP_BODY = "respbody"
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
            filterStr := ""
            switch filterPair.filterType {
                case FILTER_HOSTNAME: filterStr = p.Hostname
                case FILTER_METHOD: filterStr = p.Method
                case FILTER_PATH: filterStr = p.Path
                case FILTER_REQ_BODY: filterStr = string(p.ReqBody)
                case FILTER_STATUS: filterStr = p.Status
                case FILTER_RESP_BODY: filterStr = string(p.RespBody)
                default: {
                    slog.Warn("Unknown filter specified", "filterType", filterPair.filterType, "filterContent", filterPair.filterContent)
                }
            }

            if len(filterStr) > 0 && filterPair.negate == (strings.Contains(filterStr, filterPair.filterContent)) {
                passed = false
                break
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
