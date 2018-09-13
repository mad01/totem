package main

import (
	"fmt"
)

func main() {
	initLog(false)
	err := runCmd()
	if err != nil {
		fmt.Println(err.Error())
	}
}

// todo: adding option to limit to only one namespace with rolebinding not cluster role binding only
// todo: change internal config of users to be map since list is O(n) and map is O(1)
// todo: bug: when internaly running the service in a cluster the server address is the internal, need to use the dns name to get it to work
