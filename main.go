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

// todo: adding option to have self created rbac cluster role bindings,
// todo: adding option to limit to only one namespace with rolebinding not cluster role binding only
// todo: adding rbac setup to not show secrets as part of any view/edit/admin options
// todo: get basic user to rbac cluster role binding mapping from yaml
// todo: integrations with central auth oauth2

// todo: change internal config of users to be map since list is O(n) and map is O(1)
// todo: bug: ERRO[0048] serviceaccounts "alexander" already exists    file=http.go func=handlerKubeConfig line=55
