package golden

import (
	"flag"
	"fmt"
)

var update updateValue

func init() {
	flag.Var(&update, "update", "update golden files")
}

type updateValue string

func (u *updateValue) String() string {
	if u == nil {
		return ""
	}
	return (string)(*u)
}

func (u *updateValue) Set(v string) error {
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
