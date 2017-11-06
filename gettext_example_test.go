package gettext_test

import (
	"fmt"

	gettext "github.com/nasa9084/go-gettext"
)

func ExampleMain() {
	loc := gettext.New(`ja`, gettext.Path("test"))
	if err := loc.Load(); err != nil {
		// some error handling
	}
	fmt.Print(loc.Get(`translate this`))
}
