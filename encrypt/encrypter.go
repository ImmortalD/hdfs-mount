package encrypt

// 加密
type Encrypter interface {
	// 加密数据
	Encrypt(data []byte) []byte
	// 加密最后的数据
	EncryptLast() []byte
}

// 解密
type Decrypter interface {
	// 解密数据
	Decrypt(data []byte) []byte
}
