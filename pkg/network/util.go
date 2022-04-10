package network

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"regexp"
)

var (
	reDefaultRoute = regexp.MustCompile(`via [0-9.]+`)
)

func unusedRouteTableID() (int, error) {
	// return an unused route table ID in the range [1,252]
	for i := 1; i < 253; i = i + 1 {
		output, err := exec.Command("ip", "route", "show", "table", strconv.Itoa(i)).Output()
		if err != nil {
			return 0, err
		} else if len(output) == 0 {
			return i, nil
		}
	}
	return 0, fmt.Errorf("all routing table IDs seem to be in use")
}

func defaultRouteForTable(tableID int) (bool, string, error) {
	output, err := exec.Command("ip", "route", "show", "table", strconv.Itoa(tableID), "default").Output()
	if err != nil {
		return false, "", err
	} else if len(output) == 0 {
		return false, "", nil
	}

	parts := strings.Split(string(output[:len(output)-1]), "\n")
	switch len(parts) {
	case 0:
		return false, "", nil
	case 1:
	default:
		log.Panicf("found %d default routes for table %d", len(parts), tableID)
	}

	s := reDefaultRoute.FindString(parts[0])
	if s == "" {
		return false, "", fmt.Errorf("string didn't match regexp")
	}

	parts = strings.Split(s, " ")
	if len(parts) != 2 {
		return false, "", fmt.Errorf("unexpected regexp match")
	}
	return true, parts[1], nil
}
