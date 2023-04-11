package mysqldump

import (
	"fmt"
	"strings"
)

func GetDBNameFromDNS(dns string) (string, error) {
	ss1 := strings.Split(dns, "/")
	if len(ss1) == 2 {
		ss2 := strings.Split(ss1[1], "?")
		if len(ss2) == 2 {
			return ss2[0], nil
		}
	}

	return "", fmt.Errorf("dns error: %s", dns)
}

func ISInternalDatabase(name string) bool {
	switch name {
	case "information_schema", "mysql", "performance_schema", "test", "sys":
		return true
	default:
		return false
	}
}
