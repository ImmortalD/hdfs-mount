package openssl

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
)

const OpenSSLSaltHeader = "Salted__"

type OpenSSL struct {
	// openssl header
	openSSLSaltHeader string
	// salt
	salt           []byte
	hasHeaderWrite bool
	// 剩余header长度
	remainingHeaderLen int
	// 加密 解密
	blockMode cipher.BlockMode
	// key长度
	keyLen int
	// 缓存buffer
	inBuf        []byte
	useInBufSize int
	// 实现一个流式加解密
	// outBuf       []byte
	// wrMu         sync.Mutex
	// wCond        sync.Cond
	// rCond        sync.Cond
	// once         sync.Once // Protects closing done
	// done         chan struct{}
	// hasMoreData  chan struct{}
}

// 根据密码 salt生成key和iv
type CredsGenerator func(password []byte, keyLen int, salt []byte) (Key []byte, IV []byte)
type BlockModeFunc func(b cipher.Block, iv []byte) cipher.BlockMode

// CredsGenerator
var (
	BytesToKeyMD5    = NewBytesToKeyGenerator(md5.New)
	BytesToKeySHA1   = NewBytesToKeyGenerator(sha1.New)
	BytesToKeySHA256 = NewBytesToKeyGenerator(sha256.New)
	BytesToKeySHA384 = NewBytesToKeyGenerator(sha512.New384)
	BytesToKeySHA512 = NewBytesToKeyGenerator(sha512.New)
)

// blockMode
var (
	CBCEncrypter = cipher.NewCBCEncrypter
	CBCDecrypter = cipher.NewCBCDecrypter
)

func NewAesCipherByStrKeyAndIV(key, iv string, keyLen int, f BlockModeFunc) (*OpenSSL, error) {
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}
	ivBytes, err := hex.DecodeString(iv)
	if err != nil {
		return nil, err
	}
	return newAesCipherByKeyAndIV(keyBytes, ivBytes, keyLen, f)
}

func NewAesDencrypterByStrKeyAndIV(key, iv string, keyLen int, f BlockModeFunc) (*OpenSSL, error) {
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}
	ivBytes, err := hex.DecodeString(iv)
	if err != nil {
		return nil, err
	}
	return newAesCipherByKeyAndIV(keyBytes, ivBytes, keyLen, f)
}

func NewAesDencrypterByPass(pass, salt string, keyLen int, cg CredsGenerator, f BlockModeFunc) (*OpenSSL, error) {
	bSalt, err := hex.DecodeString(salt)
	if err != nil {
		return nil, err
	}
	return newAesEncrypterByPass(pass, bSalt, keyLen, cg, f)
}

func NewAesEncrypterByPass(pass string, keyLen int, cg CredsGenerator, f BlockModeFunc) (*OpenSSL, error) {
	salt := MustGenerateSalt()
	return newAesEncrypterByPass(pass, salt, keyLen, cg, f)
}

func newAesEncrypterByPass(pass string, salt []byte, keyLen int, cg CredsGenerator, f BlockModeFunc) (*OpenSSL, error) {
	key, iv := cg([]byte(pass), keyLen, salt)
	o, err := newAesCipherByKeyAndIV(key, iv, keyLen, f)
	if err != nil {
		return nil, err
	}
	o.salt = salt
	o.hasHeaderWrite = true
	o.remainingHeaderLen = len(o.openSSLSaltHeader) + 8
	o.openSSLSaltHeader = OpenSSLSaltHeader
	return o, nil
}

// 创建OpenSSL
func newAesCipherByKeyAndIV(key, iv []byte, keyLen int, f BlockModeFunc) (*OpenSSL, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return &OpenSSL{
		openSSLSaltHeader: OpenSSLSaltHeader,
		blockMode:         f(block, iv),
		keyLen:            keyLen,
	}, nil
}

// pkcs7Pad appends padding.
func Pkcs7Pad(data []byte, blocklen int) []byte {
	padlen := 1
	for ((len(data) + padlen) % blocklen) != 0 {
		padlen++
	}

	pad := bytes.Repeat([]byte{byte(padlen)}, padlen)
	return append(data, pad...)
}

// pkcs7Unpad returns slice of the original data without padding.
func Pkcs7Unpad(data []byte, blocklen int) ([]byte, error) {
	if blocklen <= 0 {
		return nil, fmt.Errorf("invalid blocklen %d", blocklen)
	}
	if len(data)%blocklen != 0 || len(data) == 0 {
		return nil, fmt.Errorf("invalid data len %d", len(data))
	}
	padlen := int(data[len(data)-1])
	if padlen > blocklen || padlen == 0 {
		return nil, fmt.Errorf("invalid padding")
	}
	pad := data[len(data)-padlen:]
	for i := 0; i < padlen; i++ {
		if pad[i] != byte(padlen) {
			return nil, fmt.Errorf("invalid padding")
		}
	}
	return data[:len(data)-padlen], nil
}

// GenerateSalt generates a random 8 byte salt
func GenerateSalt() ([]byte, error) {
	// Generate an 8 byte salt
	salt := make([]byte, 8)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		return nil, err
	}

	return salt, nil
}

func MustGenerateSalt() []byte {
	s, err := GenerateSalt()
	if err != nil {
		panic(err)
	}
	return s
}

func NewBytesToKeyGenerator(hashFunc func() hash.Hash) CredsGenerator {
	df := func(in []byte) []byte {
		h := hashFunc()
		h.Write(in)
		return h.Sum(nil)
	}

	return func(password []byte, keyLen int, salt []byte) (Key []byte, IV []byte) {
		var m []byte
		var prev []byte
		//  keyLen + 16bytes iv
		minLen := keyLen/8 + 16
		for len(m) < minLen {
			a := make([]byte, len(prev)+len(password)+len(salt))
			copy(a, prev)
			copy(a[len(prev):], password)
			copy(a[len(prev)+len(password):], salt)

			prev = df(a)
			m = append(m, prev...)
		}

		size := keyLen / 8
		return m[:size], m[size : size+16]
	}
}
