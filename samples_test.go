// Copyright (c) 2015 Datacratic. All rights reserved.

package jq

import (
	"bufio"
	"log"
	"os"
)

var samples = [][]byte{}

func init() {
	file, err := os.Open("./samples.json")
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		samples = append(samples, []byte(scanner.Text()))
	}

	err = scanner.Err()
	if err != nil {
		log.Fatal(err)
	}

	if len(samples) == 0 {
		log.Fatal("failed to load JSON samples")
	}
}
