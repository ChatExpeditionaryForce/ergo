// Copyright (c) 2018 Shivaram Lingamneni <slingamn@cs.stanford.edu>
// released under the MIT license

package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net"
	"strconv"
	"strings"
	"time"
)

var (
	// slingamn's own private b32 alphabet, removing 1, l, o, and 0
	B32Encoder = base32.NewEncoding("abcdefghijkmnpqrstuvwxyz23456789").WithPadding(base32.NoPadding)

	ErrInvalidCertfp = errors.New("Invalid certfp")

	ErrNoPeerCerts = errors.New("No certfp available")

	ErrNotTLS = errors.New("Connection is not TLS")
)

const (
	SecretTokenLength = 26
	MachineId         = 1 // Since there's no scaling Ergo, id is fixed at 1. Other things can have 2-127
)

var inc uint64 = 0

// slingamn, if you ever see this, i'm sorry - I just didn't want to attach what i think is redundant data to every
// message.
func GenerateMessageId() uint64 {
	inc++
	var ts = time.Now().Unix() & 0xffffffffffff
	var flake = uint64(ts << 16)
	flake |= MachineId << 10
	flake |= inc % 0x3ff
	return flake
}

func GenerateMessageIdStr() string {
	return strconv.FormatUint(GenerateMessageId(), 10)
}

// generate a secret token that cannot be brute-forced via online attacks
func GenerateSecretToken() string {
	// 128 bits of entropy are enough to resist any online attack:
	var buf [16]byte
	rand.Read(buf[:])
	// 26 ASCII characters, should be fine for most purposes
	return B32Encoder.EncodeToString(buf[:])
}

// securely check if a supplied token matches a stored token
func SecretTokensMatch(storedToken string, suppliedToken string) bool {
	// XXX fix a potential gotcha: if the stored token is uninitialized,
	// then nothing should match it, not even supplying an empty token.
	if len(storedToken) == 0 {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(storedToken), []byte(suppliedToken)) == 1
}

// generate a 256-bit secret key that can be written into a config file
func GenerateSecretKey() string {
	var buf [32]byte
	rand.Read(buf[:])
	return base64.RawURLEncoding.EncodeToString(buf[:])
}

// Normalize openssl-formatted certfp's to oragono's format
func NormalizeCertfp(certfp string) (result string, err error) {
	result = strings.ToLower(strings.Replace(certfp, ":", "", -1))
	decoded, err := hex.DecodeString(result)
	if err != nil || len(decoded) != 32 {
		return "", ErrInvalidCertfp
	}
	return
}

func GetCertFP(conn net.Conn, handshakeTimeout time.Duration) (fingerprint string, peerCerts []*x509.Certificate, err error) {
	tlsConn, isTLS := conn.(*tls.Conn)
	if !isTLS {
		return "", nil, ErrNotTLS
	}

	// ensure handshake is performed
	tlsConn.SetDeadline(time.Now().Add(handshakeTimeout))
	err = tlsConn.Handshake()
	tlsConn.SetDeadline(time.Time{})

	if err != nil {
		return "", nil, err
	}

	peerCerts = tlsConn.ConnectionState().PeerCertificates
	if len(peerCerts) < 1 {
		return "", nil, ErrNoPeerCerts
	}

	rawCert := sha256.Sum256(peerCerts[0].Raw)
	fingerprint = hex.EncodeToString(rawCert[:])

	return fingerprint, peerCerts, nil
}
