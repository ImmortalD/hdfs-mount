package openssl

func (o *OpenSSL) Decrypt(data []byte) []byte {
	if o.hasHeaderWrite {
		if o.remainingHeaderLen >= len(data) {
			o.remainingHeaderLen -= len(data)
			return data[0:0]
		} else {
			data = data[o.remainingHeaderLen:]
			o.remainingHeaderLen = 0
			o.hasHeaderWrite = false
		}
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
		return plaint
	} else {
		// buf使用的大小
		o.useInBufSize += len(data)
		// 更新buf的数据
		o.inBuf = append(o.inBuf, data...)
		return data[0:0]
	}
}
