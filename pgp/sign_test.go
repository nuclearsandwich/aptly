package pgp

import (
	"crypto/rand"
	"io"
	"io/ioutil"
	"os"
	"path"

	. "gopkg.in/check.v1"
)

// Common set of tests shared by internal & external GnuPG implementations
type SignerSuite struct {
	signer   Signer
	verifier Verifier

	clearF    *os.File
	signedF   *os.File
	cleartext []byte

	passwordFile string
}

func (s *SignerSuite) SetUpTest(c *C) {
	tempDir := c.MkDir()

	var err error
	s.clearF, err = os.Create(path.Join(tempDir, "cleartext"))
	c.Assert(err, IsNil)

	s.cleartext = make([]byte, 0, 1024)
	_, err = rand.Read(s.cleartext)
	c.Assert(err, IsNil)

	_, err = s.clearF.Write(s.cleartext)
	c.Assert(err, IsNil)

	_, err = s.clearF.Seek(0, io.SeekStart)
	c.Assert(err, IsNil)

	s.signedF, err = os.Create(path.Join(tempDir, "signed"))
	c.Assert(err, IsNil)

	s.passwordFile = path.Join(tempDir, "password")
	f, err := os.OpenFile(s.passwordFile, os.O_CREATE|os.O_WRONLY, 0600)
	c.Assert(err, IsNil)

	_, err = f.Write([]byte("verysecret"))
	c.Assert(err, IsNil)

	f.Close()

	s.signer.SetBatch(true)
}

func (s *SignerSuite) TearDownTest(c *C) {
	s.clearF.Close()
	s.signedF.Close()
}

func (s *SignerSuite) testSignDetached(c *C) {
	c.Assert(s.signer.Init(), IsNil)

	err := s.signer.DetachedSign(s.clearF.Name(), s.signedF.Name())
	c.Assert(err, IsNil)

	err = s.verifier.VerifyDetachedSignature(s.signedF, s.clearF, false)
	c.Assert(err, IsNil)
}

func (s *SignerSuite) TestSignDetachedNoPassphrase(c *C) {
	s.signer.SetKeyRing("keyrings/aptly.pub", "keyrings/aptly.sec")

	s.testSignDetached(c)
}

func (s *SignerSuite) TestSignDetachedPassphrase(c *C) {
	s.signer.SetKeyRing("keyrings/aptly_passphrase.pub", "keyrings/aptly_passphrase.sec")
	s.signer.SetPassphrase("verysecret", "")

	s.testSignDetached(c)
}

func (s *SignerSuite) TestSignDetachedPassphraseFile(c *C) {
	s.signer.SetKeyRing("keyrings/aptly_passphrase.pub", "keyrings/aptly_passphrase.sec")
	s.signer.SetPassphrase("", s.passwordFile)

	s.testSignDetached(c)
}

func (s *SignerSuite) testClearSign(c *C, expectedKey Key) {
	c.Assert(s.signer.Init(), IsNil)

	err := s.signer.ClearSign(s.clearF.Name(), s.signedF.Name())
	c.Assert(err, IsNil)

	keyInfo, err := s.verifier.VerifyClearsigned(s.signedF, false)
	c.Assert(err, IsNil)

	c.Assert(keyInfo.GoodKeys, DeepEquals, []Key{expectedKey})
	c.Assert(keyInfo.MissingKeys, DeepEquals, []Key(nil))

	_, err = s.signedF.Seek(0, io.SeekStart)
	c.Assert(err, IsNil)
	extractedF, err := s.verifier.ExtractClearsigned(s.signedF)
	c.Assert(err, IsNil)
	defer extractedF.Close()

	extracted, err := ioutil.ReadAll(extractedF)
	c.Assert(err, IsNil)

	c.Assert(extracted, DeepEquals, s.cleartext)
}

func (s *SignerSuite) TestClearSignNoPassphrase(c *C) {
	s.signer.SetKeyRing("keyrings/aptly.pub", "keyrings/aptly.sec")

	s.testClearSign(c, "21DBB89C16DB3E6D")
}

func (s *SignerSuite) TestClearSignPassphrase(c *C) {
	s.signer.SetKeyRing("keyrings/aptly_passphrase.pub", "keyrings/aptly_passphrase.sec")
	s.signer.SetPassphrase("verysecret", "")

	s.testClearSign(c, "F30E8CB9CDDE2AF8")
}

func (s *SignerSuite) TestClearSignPassphraseFile(c *C) {
	s.signer.SetKeyRing("keyrings/aptly_passphrase.pub", "keyrings/aptly_passphrase.sec")
	s.signer.SetPassphrase("", s.passwordFile)

	s.testClearSign(c, "F30E8CB9CDDE2AF8")
}
