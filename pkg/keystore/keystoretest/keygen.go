// Copyright 2014 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build ignore

// Generate opengpg keys for Application Container Keystore. Outputs to keymap.go
// and will overwrite existing files.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/coreos/rkt/Godeps/_workspace/src/golang.org/x/crypto/openpgp"
	"github.com/coreos/rkt/Godeps/_workspace/src/golang.org/x/crypto/openpgp/armor"
)

type Key struct {
	Name              string
	Fingerprint       string
	ArmoredPublicKey  string
	ArmoredPrivateKey string
}

var output = "keymap.go"

var keymapTemplate = `// Code generated by go generate.
// Source file: keygen.go
// DO NOT EDIT!

package keystoretest

var KeyMap = map[string]*KeyDetails{
{{range .}}	"{{.Name}}": &KeyDetails{
		Fingerprint: ` + "`" + `{{.Fingerprint}}` + "`" + `,
		ArmoredPublicKey: ` + "`" + `{{.ArmoredPublicKey}}` + "`" + `,
		ArmoredPrivateKey: ` + "`" + `{{.ArmoredPrivateKey}}` + "`" + `,
	},
{{end}}}
`

var names = []string{
	"example.com",
	"coreos.com",
	"example.com/app",
	"acme.com",
	"acme.com/services",
	"acme.com/services/web/nginx",
}

func main() {
	ks := make([]Key, 0)
	for _, name := range names {
		entity, err := newEntity(name)
		if err != nil {
			log.Fatal(err)
		}

		privateKeyBuf := bytes.NewBuffer(nil)
		w0, err := armor.Encode(privateKeyBuf, openpgp.PrivateKeyType, nil)
		if err != nil {
			log.Fatal(err)
		}
		if err := entity.SerializePrivate(w0, nil); err != nil {
			log.Fatal(err)
		}
		w0.Close()

		publicKeyBuf := bytes.NewBuffer(nil)
		w1, err := armor.Encode(publicKeyBuf, openpgp.PublicKeyType, nil)
		if err != nil {
			log.Fatal(err)
		}
		if err := entity.Serialize(w1); err != nil {
			log.Fatal(err)
		}
		w1.Close()

		fingerprint := fmt.Sprintf("%x", entity.PrimaryKey.Fingerprint)
		key := Key{
			Name:              name,
			Fingerprint:       fingerprint,
			ArmoredPublicKey:  publicKeyBuf.String(),
			ArmoredPrivateKey: privateKeyBuf.String(),
		}
		ks = append(ks, key)
	}
	tmpl, err := template.New("keymap").Parse(keymapTemplate)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create(output)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	err = tmpl.Execute(f, ks)
	if err != nil {
		log.Fatal(err)
	}
}

func newEntity(name string) (*openpgp.Entity, error) {
	parts := strings.Split(name, "/")
	comment := fmt.Sprintf("%s Signing Key", name)
	email := fmt.Sprintf("signer@%s", parts[0])
	entity, err := openpgp.NewEntity("signer", comment, email, nil)
	if err != nil {
		return nil, err
	}
	if err := entity.SerializePrivate(ioutil.Discard, nil); err != nil {
		return nil, err
	}
	return entity, nil
}
