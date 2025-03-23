package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic" //是我以前没见过也没想过的曲线加密库
	"crypto/rand"
	"crypto/sha256"
	"fmt"

	"github.com/btcsuite/btcutil/base58"
)

// - 结构定义
type wallet struct {
	PriKey *ecdsa.PrivateKey
	PubKey []byte
} //原注释讲了公钥，也有关于script的一点细节

func newWalletKeyPair() *wallet {
	curve := elliptic.P256()                             //椭圆加密，这里用256已经足够
	priKey, err := ecdsa.GenerateKey(curve, rand.Reader) //随机初始私钥
	if err != nil {
		fmt.Println("ecdsa.GenerateKey err:", err)
		return nil
	}

	pubKeyRaw := priKey.PublicKey

	pubKey := append(pubKeyRaw.X.Bytes(), pubKeyRaw.Y.Bytes()...) //拼接

	wallet := wallet{priKey, pubKey} //映射
	return &wallet
}

func (w *wallet) getAddress() string {

	pubKeyHash := getPubKeyHashFromPubKey(w.PubKey)

	payload := append([]byte{byte(0x00)}, pubKeyHash...)

	checksum := checkSum(payload)

	payload = append(payload, checksum...)
	address := base58.Encode(payload)
	return address
}

func getPubKeyHashFromPubKey(pubKey []byte) []byte {
	hash1 := sha256.Sum256(pubKey)
	hasher := sha256.New()
	hasher.Write(hash1[:])

	pubKeyHash := hasher.Sum(nil)

	return pubKeyHash
}

func getPubKeyHashFromAddress(address string) []byte {
	decodeInfo := base58.Decode(address)
	if len(decodeInfo) != 25 {
		fmt.Println("getPubKeyHashFromAddress, 传入地址无效")
		return nil
	}

	pubKeyHash := decodeInfo[1 : len(decodeInfo)-4] //这里是把最后checksum给去掉
	return pubKeyHash
}

func checkSum(payload []byte) []byte {
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	checksum := second[0:4]
	return checksum
}

func isValidAddress(address string) bool {
	decodeInfo := base58.Decode(address)

	if len(decodeInfo) != 25 {
		fmt.Println("isValidAddress, 传入地址长度无效")
		return false
	}

	payload := decodeInfo[:len(decodeInfo)-4]
	checksum1 := decodeInfo[len(decodeInfo)-4:] //分离
	checksum2 := checkSum(payload)
	return bytes.Equal(checksum1, checksum2)
} //这就已经算是script了
