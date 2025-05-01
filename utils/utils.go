package utils

import "fmt"

func PanicIf(err error, prefixMessage string) {

	if err != nil {
		panic(fmt.Sprintf("%s: %s", prefixMessage, err.Error()))
	}
}
