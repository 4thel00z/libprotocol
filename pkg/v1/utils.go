package v1

import "strings"

func ContainsString(needle string, haystack ...string) bool {
	for _, hay := range haystack {
		if hay == needle {
			return true
		}
	}
	return false
}

func ContainsCaseInsensitiveString(needle string, haystack ...string) bool {
	needle = strings.ToLower(needle)
	for _, hay := range haystack {
		if strings.ToLower(hay) == needle {
			return true
		}
	}
	return false
}

func Any(truths ...bool) bool {
	for _, truth := range truths {
		if truth {
			return true
		}
	}
	return false
}

func All(truths ...bool) bool {
	for _, truth := range truths {
		if !truth {
			return false
		}
	}
	return true
}
