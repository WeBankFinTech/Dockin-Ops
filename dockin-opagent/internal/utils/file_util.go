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

package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}

	return s.IsDir()
}

func IsFile(path string) bool {
	return !IsDir(path)
}

func GetPrefix(file string) string {
	// tar strips the leading '/' if it's there, so we will too
	return strings.TrimLeft(file, "\\")
}

func StripPathShortcuts(p string) string {
	newPath := path.Clean(p)
	trimmed := strings.TrimPrefix(newPath, "../")

	for trimmed != newPath {
		newPath = trimmed
		trimmed = strings.TrimPrefix(newPath, "../")
	}

	// trim leftover {".", ".."}
	if newPath == "." || newPath == ".." {
		newPath = ""
	}

	if len(newPath) > 0 && string(newPath[0]) == "/" {
		return newPath[1:]
	}

	return newPath
}

func GetFileMd5(file string) string {
	f, _ := os.Open(file)
	defer f.Close()
	md5 := md5.New()
	io.Copy(md5, f)
	return hex.EncodeToString(md5.Sum(nil))
}

func GetFileHash(file string) string {
	f, _ := os.Open(file)
	defer f.Close()
	h := sha1.New()
	io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil))
}

func getPrefix(file string) string {
	// tar strips the leading '/' if it's there, so we will too
	return strings.TrimLeft(file, "/")
}

func linkJoin(base, link string) string {
	if filepath.IsAbs(link) {
		return link
	}
	return filepath.Join(base, link)
}

func substr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}

func getParentDirectory(directory string) string {
	return substr(directory, 0, strings.LastIndex(directory, "/"))
}

func FileSplit(filePath, destPath string, chunkSize int64) ([]string, error) {
	var (
		fileList []string
	)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	num := int64(math.Ceil(float64(fileInfo.Size()) / float64(chunkSize)))

	/*
		if num <= 1 {
			return append(fileList, filePath), nil
		}
	*/
	if !Exists(destPath) {
		os.MkdirAll(destPath, os.ModePerm)
	}

	fi, err := os.OpenFile(filePath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}

	b := make([]byte, chunkSize)
	var i int64 = 1
	for ; i <= num; i++ {
		fi.Seek((i-1)*(chunkSize), 0)

		if len(b) > int((fileInfo.Size() - (i-1)*chunkSize)) {
			b = make([]byte, fileInfo.Size()-(i-1)*chunkSize)
		}

		fi.Read(b)

		str := "_" + fmt.Sprintf("%04d", i)
		//name,suffix := util.FileSplitFileNameSuffix(filePath)
		newfilename := path.Join(destPath, path.Base(filePath)+str)
		//newfilename := name + str + suffix
		f, err := os.OpenFile(newfilename, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			fi.Close()
			return fileList, err
		}

		fileList = append(fileList, newfilename)
		f.Write(b)
		f.Close()
	}
	fi.Close()

	return fileList, nil
}
