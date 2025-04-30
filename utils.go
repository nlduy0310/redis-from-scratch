package main

import "fmt"

func panicIf(err error, prefixMessage string) {
	if err != nil {
		fmt.Errorf(fmt.Sprintf("%s: :v", prefixMessage, err))
	}
}
