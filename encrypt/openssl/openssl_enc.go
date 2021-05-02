package openssl

func (o *OpenSSL) Encrypt(data []byte) []byte {
	if data == nil || len(data) == 0 {
		return data
	}

	res := make([]byte, 0)
	if o.hasHeaderWrite {
		res = append([]byte(o.openSSLSaltHeader), o.salt...)
		o.hasHeaderWrite = false
	}

	if o.useInBufSize+len(data) >= o.blockMode.BlockSize() {
		// 剩余的未加密的数据长度
		overLen := (o.useInBufSize + len(data)) % o.blockMode.BlockSize()
		// 需要加密的数据
		plaint := append(o.inBuf[0:o.useInBufSize], data[:(len(data)-overLen)]...)
		// 加密数据
		o.blockMode.CryptBlocks(plaint, plaint)
		// buf使用的大小
		o.useInBufSize = overLen
		// 更新buf的数据
		o.inBuf = data[len(data)-overLen:]
		res = append(res, plaint...)
	} else {
		// buf使用的大小
		o.useInBufSize += len(data)
		// 更新buf的数据
		o.inBuf = append(o.inBuf, data...)
		// res = append(res, data[0:0]...)
	}

	return res
}

func (o *OpenSSL) EncryptLast() []byte {
	res := make([]byte, 0)
	if o.hasHeaderWrite {
		res = append([]byte(o.openSSLSaltHeader), o.salt...)
		o.hasHeaderWrite = false
	}

	lastData := Pkcs7Pad(o.inBuf, o.blockMode.BlockSize())
	o.blockMode.CryptBlocks(lastData, lastData)
	return append(res, lastData...)
}
