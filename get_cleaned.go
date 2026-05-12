package main

import (
	"strings"
)

func getCleanedBody(cB string) string {
	splitCB := strings.Split(cB, " ")
	for i, word := range splitCB {
		lc := strings.ToLower(word)
		if lc == "kerfuffle" || lc == "sharbert" || lc == "fornax" {
			splitCB[i] = "****"
		}
	}
	joinedCB := strings.Join(splitCB, " ")
	return joinedCB
}
