package output

import "os"

func CommandName() string {
	if os.Args[0] == "kubectl" || os.Args[0] == "oc" {
		return os.Args[0] + " rhmctl"
	}

	return "rhmctl"
}
