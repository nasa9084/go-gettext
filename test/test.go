package main

import (
	"fmt"

	gettext "github.com/nasa9084/go-gettext"
)

func main() {
	loc := gettext.New("ja")
	if err := loc.Parse(); err != nil {
		panic(err)
	}
	// TRANSLATORS: sample comment
	fmt.Println(loc.Get(`translate this`))
}
