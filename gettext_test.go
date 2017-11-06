package gettext_test

import (
	"testing"

	gettext "github.com/nasa9084/go-gettext"
)

func TestGet(t *testing.T) {
	in := `translate this`
	expect := "これをほんやくしてください"
	loc := gettext.New("ja", gettext.Path("test"))
	if err := loc.Load(); err != nil {
		t.Error(err)
		return
	}
	if loc.Get(in) != expect {
		t.Errorf(`"%s" != "%s"`, in, expect)
		return
	}
}
