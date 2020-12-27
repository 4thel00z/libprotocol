package v1

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"math/big"
	"strings"
)

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

func Send(obj interface{}, output chan<- []byte) error {
	payload, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	output <- payload
	return nil
}

func Max(array []float64) (float64, error) {
	if len(array) == 0 {
		return -1, errors.New("len(array) == 0")
	}
	max := array[0]
	for _, value := range array {
		if max < value {
			max = value
		}
	}
	return max, nil
}

func Min(array []float64) (float64, error) {
	if len(array) == 0 {
		return -1, errors.New("len(array) == 0")
	}
	min := array[0]
	for _, value := range array {
		if min > value {
			min = value
		}
	}
	return min, nil
}

func InitialSequenceNumber(max int64) int64 {
	n := big.NewInt(max)
	b, _ := rand.Int(rand.Reader, n)
	return b.Int64()
}