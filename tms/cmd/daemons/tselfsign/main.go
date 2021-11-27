// tselfsign generates certificates.
package main

// https://golang.org/src/crypto/tls/generate_cert.go

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	certFile       string
	keyFile        string
	ipAddressesArg string
	dnsNamesArg    string
	caFile         string
	genMC          bool
	mCrts          string
	caKey          string
	force          bool
	genCA          bool
)

func init() {
	flag.StringVar(&certFile, "certificate", "/etc/trident/certificate.pem",
		"write out certificate to this file")
	flag.StringVar(&keyFile, "key", "/etc/trident/key.pem",
		"write out key to this file")
	flag.StringVar(&ipAddressesArg, "ip", "",
		"comma delimited list of IP addresses to include")
	flag.StringVar(&dnsNamesArg, "dns", "",
		"comma delimted list of DNS names to include")
	flag.StringVar(&caFile, "CACertificate", "/etc/trident/mongoCA.crt",
		"write out CA file to the referenced file")
	flag.StringVar(&caKey, "CAKey", "/etc/trident/mongoCA.key",
		"write out CA PK to referenced file")
	flag.BoolVar(&genMC, "generateMongoCertificate", false,
		"flag to generate mongo certs default is false")
	flag.StringVar(&mCrts, "mongoCertificatePath", "/etc/trident/",
		"write out mongo pem file to the referenced file")
	flag.BoolVar(&force, "f", false,
		"force to regenerate CA key and crt even if they already exist")
	flag.BoolVar(&genCA, "generateCA", false,
		"flag to generate CA crt and key default is false")
}

func main() {
	flag.Parse()

	ipAddresses := []net.IP{}
	for _, ipAddress := range strings.Split(ipAddressesArg, ",") {
		ipAddress = strings.TrimSpace(ipAddress)
		if len(ipAddress) > 0 {
			ipAddresses = append(ipAddresses, net.ParseIP(ipAddress))
		}
	}
	dnsNames := []string{}
	for _, dnsName := range strings.Split(dnsNamesArg, ",") {
		dnsName = strings.TrimSpace(dnsName)
		if len(dnsName) > 0 {
			dnsNames = append(dnsNames, dnsName)
		}
	}

	if genCA || genMC {

		if genCA {
			_, kerr := os.Stat(caKey)
			_, caerr := os.Stat(caFile)
			if os.IsNotExist(kerr) || os.IsNotExist(caerr) || force {
				//generate CA private key
				cmdCAKey := "sudo openssl genrsa -out " + strings.Split(caKey, " ")[0] + " -passout pass:orolia -aes256 8192"
				gck := exec.Command("/bin/sh", "-c", cmdCAKey)
				if out, err := gck.CombinedOutput(); err != nil {
					log.Fatal(fmt.Sprintf("Error: %s\n", err.Error()))

				} else {
					log.Print(string(out))
				}

				//generate CA crt file
				cmdCACrt := "sudo openssl req -x509 -new -extensions v3_ca -key " + strings.Split(caKey, " ")[0] + " -passin pass:orolia -subj \"/O=Orolia/OU=trident/CN=GroundZero\" -days 365 -out " + strings.Split(caFile, " ")[0]
				gca := exec.Command("/bin/sh", "-c", cmdCACrt)
				if out, err := gca.CombinedOutput(); err != nil {
					log.Fatal(fmt.Sprintf("Error: %s\n", err.Error()))

				} else {
					log.Print(string(out) + " " + strings.Split(caFile, " ")[0] + " successfuly created")
				}
			} else {
				log.Fatal(fmt.Sprintf("Error: %s or %s already exist use -f to force \n", caKey, caFile))
			}
		}

		if genMC {
			if len(ipAddresses) == 0 {
				log.Fatal("No ip address available to inject into mongo pem subject")
			}
			_, merr := os.Stat(mCrts + "mongo.pem")
			if os.IsNotExist(merr) || force {
				//generate mongo key and crs
				cmdMongoPem := "sudo openssl req -new -nodes -newkey rsa:4096 -subj \"/O=Orolia/OU=trident/CN=" + ipAddresses[0].String() + "\" -keyout " + mCrts + "mongo.key -out " + mCrts + "mongo.csr"
				gmc := exec.Command("/bin/sh", "-c", cmdMongoPem)
				if out, err := gmc.CombinedOutput(); err != nil {
					log.Fatal(fmt.Sprintf("Error: %s\n", err.Error()))

				} else {
					log.Print(string(out))
				}

				//sign the crs
				cmdCrsSign := "sudo openssl x509 -CA " + strings.Split(caFile, " ")[0] + " -CAkey " + strings.Split(caKey, " ")[0] + " -passin pass:orolia -CAcreateserial -req -days 365 -in " + mCrts + "mongo.csr" + " -out " + mCrts + "mongo.crt"
				gcs := exec.Command("/bin/sh", "-c", cmdCrsSign)
				if out, err := gcs.CombinedOutput(); err != nil {
					log.Fatal(fmt.Sprintf("Error: %s\n", err.Error()))

				} else {
					log.Print(string(out))
				}

				cmdRmCsr := "sudo rm " + mCrts + "mongo.csr"
				grc := exec.Command("/bin/sh", "-c", cmdRmCsr)
				if _, err := grc.CombinedOutput(); err != nil {
					log.Fatal(fmt.Sprintf("Error: %s\n", err.Error()))

				}

				//construct the pem file from key and crt
				cmdPemFile := "sudo cat " + mCrts + "mongo.key " + mCrts + "mongo.crt > " + mCrts + "mongo.pem"
				gpf := exec.Command("/bin/sh", "-c", cmdPemFile)
				if _, err := gpf.CombinedOutput(); err != nil {
					log.Fatal(fmt.Sprintf("Error: %s\n", err.Error()))

				}

			} else {
				log.Fatal(fmt.Sprintf("Error: %s/mongo.pem already exist use -f to force \n", mCrts))
			}
		}
		os.Exit(0)
	}

	if len(ipAddresses) == 0 && len(dnsNames) == 0 {
		dnsNames = []string{"localhost"}
	}

	name := ""
	if len(dnsNames) > 0 {
		name = dnsNames[0]
	} else {
		name = ipAddresses[0].String()
	}

	const rsaBits = 2048
	priv, err := rsa.GenerateKey(rand.Reader, rsaBits)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	notBefore := time.Now()
	notAfter := time.Now().Add(100 * 365 * 24 * time.Hour)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Orolia Prisma C2 (" + name + ")"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,

		IPAddresses: ipAddresses,
		DNSNames:    dnsNames,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatalf("failed to create certificate: %s", err)
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		log.Fatalf("failed to open %s for writing: %s", certFile, err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("failed to open %s for writing: %s", keyFile, err)
	}
	pem.Encode(keyOut, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})
	keyOut.Close()
}
