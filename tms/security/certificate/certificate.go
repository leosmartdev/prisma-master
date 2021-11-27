// Package certificate provides functionality to implement hitless TLS cert rotation
package certificate

import (
	"crypto/tls"
	"prisma/gogroup"
	"prisma/tms/log"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// WrappedCertificate struct protects the certificate using a mutex.
type WrappedCertificate struct {
	sync.Mutex
	certificate *tls.Certificate
}

// GetCertificate return the certificate value from WrappedCertificate
func (c *WrappedCertificate) GetCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	c.Lock()
	defer c.Unlock()

	return c.certificate, nil
}

// LoadCertificate generate X509 keypair cert then assigns it to WrappedCertificate.
func (c *WrappedCertificate) LoadCertificate(certFile, keyFile string) error {
	c.Lock()
	defer c.Unlock()
	// Only update the cert and key if the load is successful
	certAndKey, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	c.certificate = &certAndKey
	return err
}

// WatchCertificateFiles watches for certificate and key files for create or write operations to call load certificate.
func (c *WrappedCertificate) WatchCertificateFiles(ctxt gogroup.GoGroup, watcher *fsnotify.Watcher, certFile, keyFile string) {
	for {
		select {
		case event := <-watcher.Events:
			if event.Name == certFile || event.Name == keyFile {
				log.Info("%+v file operation on: %+v", event.Op, event.Name)
				// Only take action if keyfile or certfile are being created or written into.
				if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
					err := c.LoadCertificate(certFile, keyFile)
					if err != nil {
						log.Error("Could not load new certificate key pair: %+v", err)
					}
				}
			}
		case err := <-watcher.Errors:
			log.Error("Cert watched error: %+v", err)
		}
	}
}
