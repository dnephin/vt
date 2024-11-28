package golden

import (
	"flag"
	"fmt"
	"os"
)

var update updateValue

func init() {
	flag.Var(&update, "update", "update golden files")
}

type updateValue string

func (u *updateValue) IsBoolFlag() bool {
	return true
}

func (u *updateValue) String() string {
	if u == nil {
		return ""
	}
	return (string)(*u)
}

func (u *updateValue) Set(v string) error {
	if v == "true" {
		v = os.Getenv("GOLDEN_UPDATE")
	}
	switch v {
	case "no":
		// used internally for testing
	case "always", "yes", "force":
		*u = "yes"
	default:
		*u = updateValue(v)
	}
	return nil
}

func (u *updateValue) Requested(gotHash string) bool {
	if u == nil || *u == "" {
		return false
	}
	if *u == "yes" {
		return true
	}
	if string(*u) == gotHash {
		return true
	}

	fmt.Printf("Refusing to update because the hash %v did not match the actual value hash %v\n", *u, gotHash)
	return false
}

// Get provides compatibility with other libraries that define
// an optional bool flag for -update.
func (u *updateValue) Get() any {
	return u.Requested("")
}
