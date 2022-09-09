package msgo

import (
	"strings"
)

func SubStringLast(url, groupName string) string {
	index := strings.Index(url, groupName)
	if index == -1 {
		return ""
	}

	return url[index+len(groupName):]
}
