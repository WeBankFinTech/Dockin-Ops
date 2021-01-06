/*
 * Copyright (C) @2021 Webank Group Holding Limited
 * <p>
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * <p>
 * http://www.apache.org/licenses/LICENSE-2.0
 * <p>
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 */

package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

type Aes struct {
	block cipher.Block
}

func NewAes(key string) (*Aes, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	return &Aes{
		block:block,
	}, nil
}

func (a *Aes)AesEncrypt(plaintext string) (string, error) {
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	cipher.NewCFBEncrypter(a.block, iv).XORKeyStream(ciphertext[aes.BlockSize:],
		[]byte(plaintext))
	return hex.EncodeToString(ciphertext), nil
}

func (a *Aes)AesDecrypt(d string) (string, error) {
	ciphertext, err := hex.DecodeString(d)
	if err != nil {
		return "", err
	}
	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	fmt.Println(len(ciphertext), len(iv))
	cipher.NewCFBDecrypter(a.block, iv).XORKeyStream(ciphertext, ciphertext)
	return string(ciphertext), nil
}