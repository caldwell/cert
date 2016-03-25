// -*- compile-command: "env GOPATH=`pwd`/.go go build new-csr.go"; -*-
package main

import (
//	"bufio"
//	"bytes"
//	"encoding/json"
        "encoding/asn1"
        "encoding/pem"
//	"errors"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"github.com/docopt/docopt-go"
	"github.com/caldwell/docopt-unmarshall"
	// "io"
	"log"
//	"net/http"
//	"net/url"
	"os"
//	"os/exec"
	"path"
//	"regexp"
//	"reflect"
	"strconv"
//	"strings"
//	"time"
)

var verbose = 0

func StrPtr(s interface{}) *string {
	switch str := s.(type) { case string: return &str }
	return nil
}

func _try(s interface{}, err error) interface{} {
	if err != nil {
		log.Output(2, fmt.Sprintf("Fatal: %s\n", err.Error()))
		os.Exit(1)
	}
	return s
}

type basicConstraints struct {
	IsCA       bool `asn1:"optional"`
	MaxPathLen int  `asn1:"optional,default:-1"`
}
var ( // lifted from https://golang.org/src/crypto/x509/x509.go
	oidExtensionSubjectKeyId          = []int{2, 5, 29, 14}
	oidExtensionKeyUsage              = []int{2, 5, 29, 15}
	oidExtensionExtendedKeyUsage      = []int{2, 5, 29, 37}
	oidExtensionAuthorityKeyId        = []int{2, 5, 29, 35}
	oidExtensionBasicConstraints      = []int{2, 5, 29, 19}
	oidExtensionSubjectAltName        = []int{2, 5, 29, 17}
	oidExtensionCertificatePolicies   = []int{2, 5, 29, 32}
	oidExtensionNameConstraints       = []int{2, 5, 29, 30}
	oidExtensionCRLDistributionPoints = []int{2, 5, 29, 31}
	oidExtensionAuthorityInfoAccess   = []int{1, 3, 6, 1, 5, 5, 7, 1, 1}
)

func main() {
	me := path.Base(os.Args[0])
	options := struct {
		Help           bool   `docopt:"--help"`
		Verbose        int    `docopt:"--verbose"`
		ExpireDays     int    `docopt:"--days"`
		RsaBits        int    `docopt:"--bits"`
		Hash           string `docopt:"--hash"`
		Country        string `docopt:"--country"`
		State          string `docopt:"--state"`
		Locality       string `docopt:"--locality"`
		Organization   string `docopt:"--organization"`
		Section        string `docopt:"--section"`
		Cn             string `docopt:"--cn"`
		Email          string `docopt:"--email"`
		AltDns       []string `docopt:"--alt-dns"`
		OutKey         string `docopt:"<out_key>"`
		OutCsr         string `docopt:"<out_csr>"`
	}{
		Verbose:       0,
		ExpireDays:    100,
		RsaBits:       4096,
		Hash:          "sha256",
		Country:       "US",
		State:         "Denial",
		Locality:      "Close",
		Organization:  "Pretty Good",
		Section:       "9",
		Email:         "webmaster@example.com",
	}
	usage := `
Usage:
  `+me+` [options] [(-v | --verbose)...] [--alt-dns=<alt>...] --cn=<cn> <out_key> <out_csr>

Options:
   -v --verbose                   Turn up the verbosity
   -d --days=<expire>             Days until certificate expires (default: `+strconv.Itoa(options.ExpireDays)+`)
   -b --bits=<rsa_bits>           Number of bits in new RSA key (default: `+strconv.Itoa(options.RsaBits)+`)
   -h --hash=<hash_algorithm>     Which hashing algorithm to use (default: `+options.Hash+`)
   --country=<country>            Country (default: `+options.Country+`)
   --state=<state>                State (default: `+options.State+`)
   --locality=<locality>          Locality (default: `+options.Locality+`)
   --organization=<organization>  Organization (default: `+options.Organization+`)
   --section=<section>            Section (default: `+options.Section+`)
   --cn=<cn>                      Common Name (usually a domain name)
   --email=<email>                Email (default: `+options.Email+`)
   --alt-dns=<alt>                Alt DNS name (can be specified multiple times)
   --help                         Show this message
`

	log.SetFlags(0)
	arguments, err := docopt.Parse(usage, nil, true, me, false)
	if err != nil { log.Fatal("docopt: ", err) }
	err = docopt_unmarshall.DocoptUnmarshall(arguments, &options)
	if err != nil { log.Fatal("options: ", err) }
	verbose = options.Verbose

	if verbose > 2 { fmt.Printf("options: %v\n", options) }

	key, err := rsa.GenerateKey(rand.Reader, options.RsaBits)
	if err != nil { log.Fatal("rsa.GenerateKey: ", err) }

	csr, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{
		SignatureAlgorithm: x509.SHA256WithRSA,//map[string]SignatureAlgorithm{ "sha256": x509.SHA256WithRSA, "sha1": x509.SHA1WithRSA, "sha384": x509.SHA384WithRSA, "sha512": x509:SHA512WithRSA }[options.Hash],
		PublicKeyAlgorithm: x509.RSA,
		PublicKey:          &key.PublicKey,
		Subject:            pkix.Name{
			CommonName:		options.Cn,
			Country:		[]string{options.Country},
			Organization:		[]string{options.Organization},
			OrganizationalUnit:	[]string{options.Section},
			Locality:		[]string{options.Locality},
			Province:		[]string{options.State},
		},
		DNSNames:           append([]string{options.Cn}, options.AltDns...),
		ExtraExtensions: []pkix.Extension{
			pkix.Extension{
				Id: oidExtensionBasicConstraints,
				Value: _try(asn1.Marshal(basicConstraints{IsCA: false, MaxPathLen: -1})).([]byte),
			},
			// Missing these, do we really need them?
			// csr.add_extension extension_factory.create_extension('keyUsage', 'keyEncipherment,dataEncipherment,digitalSignature', true)
			// csr.add_extension extension_factory.create_extension('subjectKeyIdentifier', 'hash')
			// csr.add_extension extension_factory.create_extension('authorityKeyIdentifier','keyid,issuer')
			// pkix.Extension{
			// 	Id: 
			// },
		},
	}, key)
	if err != nil { log.Fatal("x509.CreateCertificateRequest: ", err) }
	csr_file, err := os.OpenFile(options.OutCsr, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil { log.Fatal(options.OutCsr, ": ", err) }
	defer csr_file.Close()
	pem.Encode(csr_file, &pem.Block{ Type: "CERTIFICATE REQUEST", Bytes: csr })

	key_file, err := os.OpenFile(options.OutKey, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil { log.Fatal(options.OutKey, ": ", err) }
	defer key_file.Close()
	pem.Encode(key_file, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
}
