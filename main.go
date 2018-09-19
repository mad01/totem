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
