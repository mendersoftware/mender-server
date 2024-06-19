// Copyright 2023 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mendersoftware/mender-server/pkg/identity"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatalf("usage: dumpidentity <token>")
	}

	token := os.Args[1]
	idata, err := identity.ExtractIdentity(token)
	if err != nil {
		log.Fatalf("error: %v", err)
		return
	}

	fmt.Printf("%+v\n", idata)
}
