package openssl

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestNewAesEncryptByPass(t *testing.T) {
	enc, err := NewAesEncrypterByPass("1", 256, BytesToKeySHA512, CBCEncrypter)
	if err != nil {
		t.Fatalf("NewAesEncrypterByPass error:%v\n", err)
	}

	buf := bytes.NewBufferString("")
	plaintBuf := bytes.NewBufferString("")

	for i := 0; i < 200; i++ {
		data := fmt.Sprintf("%d", i)
		e := enc.Encrypt([]byte(data))
		buf.Write(e)
		plaintBuf.WriteString(data)
	}
	buf.Write(enc.EncryptLast())
	i := buf.Bytes()
	v := i[len(OpenSSLSaltHeader) : len(OpenSSLSaltHeader)+8]

	dec, err := NewAesDencrypterByPass("1", hex.EncodeToString(v), 256, BytesToKeySHA512, CBCDecrypter)
	if err != nil {
		t.Fatalf("NewAesEncrypterByPass error:%v\n", err)
	}
	decBuf := bytes.NewBufferString("")
	unpad, err := Pkcs7Unpad(dec.Decrypt(i), dec.blockMode.BlockSize())
	if err != nil {
		t.Fatalf("NewAesEncrypterByPass error:%v\n", err)
	}
	decBuf.Write(unpad)

	if !bytes.Equal(unpad, plaintBuf.Bytes()) {
		t.Fatalf("Enc and Dec not equal")
	}

}

func TestNewAesCipherByStrKeyAndIV(t *testing.T) {
	enc, err := NewAesCipherByStrKeyAndIV("27a9d638039f3bf2b1b2fb05ecedbbc4e585efcb9d429406d2b8341cd253a6eb",
		"d256df4f073490b705223748d170d117", 256, CBCEncrypter)

	if err != nil {
		t.Fatalf("NewAesEncrypterByPass error:%v\n", err)
	}

	buf := bytes.NewBufferString("")
	plaintBuf := bytes.NewBufferString("")

	for i := 0; i < 200; i++ {
		data := fmt.Sprintf("%d", i)
		e := enc.Encrypt([]byte(data))
		buf.Write(e)
		plaintBuf.WriteString(data)
	}
	buf.Write(enc.EncryptLast())
	i := buf.Bytes()

	dec, err := NewAesDencrypterByStrKeyAndIV("27a9d638039f3bf2b1b2fb05ecedbbc4e585efcb9d429406d2b8341cd253a6eb",
		"d256df4f073490b705223748d170d117", 256, CBCDecrypter)
	if err != nil {
		t.Fatalf("NewAesEncrypterByPass error:%v\n", err)
	}
	decBuf := bytes.NewBufferString("")
	unpad, err := Pkcs7Unpad(dec.Decrypt(i), dec.blockMode.BlockSize())
	if err != nil {
		t.Fatalf("NewAesEncrypterByPass error:%v\n", err)
	}
	decBuf.Write(unpad)

	if !bytes.Equal(unpad, plaintBuf.Bytes()) {
		t.Fatalf("Enc and Dec not equal")
	}

}
