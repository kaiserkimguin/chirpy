package main

import (
	"errors"
	"strings"
)

func getCleanedBody(cB string) (string, error) {
	splitCB := strings.Split(cB, " ")
	for i, word := range splitCB {
		lc := strings.ToLower(word)
		if lc == "kerfuffle" || lc == "sharbert" || lc == "fornax" {
			splitCB[i] = "****"
		}
	}
	joinedCB := strings.Join(splitCB, " ")
	if len(joinedCB) > 140 {
		return "", errors.New("bad request")
	} else {
		return joinedCB, nil
	}
}
