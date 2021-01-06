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
	"archive/tar"
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/webankfintech/dockin-opsctl/internal/log"

	"github.com/klauspost/compress/gzip"
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
	return strings.TrimLeft(file, "\\")
}

func StripPathShortcuts(p string) string {
	newPath := path.Clean(p)
	trimmed := strings.TrimPrefix(newPath, "../")

	for trimmed != newPath {
		newPath = trimmed
		trimmed = strings.TrimPrefix(newPath, "../")
	}

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

func MakeGzTar(srcPath, dstPath, tarFile string) error {

	if !Exists(tarFile) {
		os.Remove(tarFile)
	}

	fw, err := os.Create(tarFile)
	if err != nil {
		return err
	}
	defer fw.Close()

	bio := bufio.NewWriter(fw)
	defer bio.Flush()
	return MakeTar(srcPath, dstPath, bio)
}

func MakeTar(srcPath, destPath string, writer io.Writer) error {
	tarWriter := tar.NewWriter(writer)
	defer tarWriter.Close()

	srcPath = path.Clean(srcPath)
	destPath = path.Clean(destPath)
	return recursiveTar(path.Dir(srcPath), path.Base(srcPath), path.Dir(destPath), path.Base(destPath), tarWriter)
}

func MakeTarV2(srcPath string, writer io.Writer) error {

	tarWriter := tar.NewWriter(writer)
	defer tarWriter.Close()

	srcPath = path.Clean(srcPath)
	return recursiveTar(filepath.Dir(srcPath), filepath.Base(srcPath), filepath.Dir(srcPath), filepath.Base(srcPath), tarWriter)
}

func recursiveTar(srcBase, srcFile, destBase, destFile string, tw *tar.Writer) error {
	srcPath := path.Join(srcBase, srcFile)
	matchedPaths, err := filepath.Glob(srcPath)
	if err != nil {
		return err
	}
	for _, fpath := range matchedPaths {
		stat, err := os.Lstat(fpath)
		if err != nil {
			return err
		}
		if stat.IsDir() {
			files, err := ioutil.ReadDir(fpath)
			if err != nil {
				return err
			}
			if len(files) == 0 {
				//case empty directory
				hdr, _ := tar.FileInfoHeader(stat, fpath)
				hdr.Name = destFile
				if err := tw.WriteHeader(hdr); err != nil {
					return err
				}
			}
			for _, f := range files {
				if err := recursiveTar(srcBase, path.Join(srcFile, f.Name()), destBase, path.Join(destFile, f.Name()), tw); err != nil {
					return err
				}
			}
			return nil
		} else if stat.Mode()&os.ModeSymlink != 0 {
			//case soft link
			hdr, _ := tar.FileInfoHeader(stat, fpath)
			target, err := os.Readlink(fpath)
			if err != nil {
				return err
			}

			hdr.Linkname = target
			hdr.Name = destFile
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
		} else {
			//case regular file or other file type like pipe
			hdr, err := tar.FileInfoHeader(stat, fpath)
			if err != nil {
				return err
			}
			hdr.Name = destFile

			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}

			f, err := os.Open(fpath)
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
			return f.Close()
		}
	}
	return nil
}

func UntarGzAll(reader io.Reader, destDir, prefix string) {
	gzReader, _ := gzip.NewReader(reader)
	UntarAll(gzReader, destDir, prefix)
}

func UntarAll(reader io.Reader, destDir, prefix string) error {
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err != io.EOF {
				return err
			}

			break
		}

		if !strings.HasPrefix(header.Name, prefix) {
			return fmt.Errorf("tar contents corrupted")
		}

		mode := header.FileInfo().Mode()
		destFileName := path.Join(destDir, header.Name[len(prefix):])
		baseName := path.Dir(destFileName)

		if err := os.MkdirAll(baseName, 0755); err != nil {
			return err
		}
		if header.FileInfo().IsDir() {
			if err := os.MkdirAll(destFileName, 0755); err != nil {
				return err
			}
			continue
		}

		dir, file := filepath.Split(destFileName)
		evaledPath, err := filepath.EvalSymlinks(dir)
		if err != nil {
			return err
		}
		if !isDestRelative(destDir, destFileName) || !isDestRelative(destDir, filepath.Join(evaledPath, file)) {
			//fmt.Fprintf(o.IOStreams.ErrOut, "warning: link %q is pointing to %q which is outside target destination, skipping\n", destFileName, header.Linkname)
			continue
		}

		if mode&os.ModeSymlink != 0 {
			linkname := header.Linkname
			if !isDestRelative(destDir, linkJoin(destFileName, linkname)) {
				//fmt.Fprintf(o.IOStreams.ErrOut, "warning: link %q is pointing to %q which is outside target destination, skipping\n", destFileName, header.Linkname)
				continue
			}
			if err := os.Symlink(linkname, destFileName); err != nil {
				return err
			}
		} else {
			outFile, err := os.Create(destFileName)
			if err != nil {
				return err
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}
			if err := outFile.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}

func isDestRelative(base, dest string) bool {
	fullPath := dest
	if !filepath.IsAbs(dest) {
		fullPath = filepath.Join(base, dest)
	}
	relative, err := filepath.Rel(base, fullPath)
	if err != nil {
		return false
	}
	return relative == "." || relative == stripPathShortcuts(relative)
}

func getPrefix(file string) string {

	return strings.TrimLeft(file, "/")
}

func linkJoin(base, link string) string {
	if filepath.IsAbs(link) {
		return link
	}
	return filepath.Join(base, link)
}

func stripPathShortcuts(p string) string {
	newPath := path.Clean(p)
	trimmed := strings.TrimPrefix(newPath, "../")

	for trimmed != newPath {
		newPath = trimmed
		trimmed = strings.TrimPrefix(newPath, "../")
	}

	if newPath == "." || newPath == ".." {
		newPath = ""
	}

	if len(newPath) > 0 && string(newPath[0]) == "/" {
		return newPath[1:]
	}

	return newPath
}

func Validate(path string) error {
	var dir string
	if path[len(path)-1:] == "/" {
		dir = path
	} else {
		dir = getParentDirectory(path)
	}
	return os.MkdirAll(dir, os.ModeDir)
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

func UntarFile(reader io.Reader, dest string) error {
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Debugf("ERROR: cannot read tar file, error=[%v]\n", err)
			return err
		}

		_, err = json.Marshal(header)
		if err != nil {
			log.Debugf("ERROR: cannot parse header, error=[%v]\n", err)
			return err
		}

		info := header.FileInfo()
		if info.IsDir() {
			filePath := filepath.Join(dest, header.Name)
			if err = os.MkdirAll(filePath, 0755); err != nil {
				log.Debugf("ERROR: cannot mkdir file, error=[%v]\n", err)
				return err
			}
		} else {
			filePath := filepath.Join(dest, path.Dir(header.Name))
			if err = os.MkdirAll(filePath, 0755); err != nil {
				log.Debugf("ERROR: cannot file mkdir file, error=[%v]\n", err)
				return err
			}

			filePath = filepath.Join(dest, header.Name)
			file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
			if err != nil {
				log.Debugf("ERROR: cannot open file, error=[%v]\n", err)
				return err
			}
			defer file.Close()

			_, err = io.Copy(file, tarReader)
			if err != nil {
				log.Debugf("ERROR: cannot write file, error=[%v]\n", err)
				return err
			}
		}
	}
	return nil
}

func tarFile(src, dst string) (err error) {
	fw, err := os.Create(dst)
	if err != nil {
		return
	}
	defer fw.Close()

	tw := tar.NewWriter(fw)

	defer tw.Close()

	return filepath.Walk(src, func(fileName string, fi os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		hdr, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}

		hdr.Name = strings.TrimPrefix(fileName, string(filepath.Separator))
		hdr.Name = filepath.Base(hdr.Name)

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		fr, err := os.Open(fileName)
		defer fr.Close()
		if err != nil {
			return err
		}

		n, err := io.Copy(tw, fr)
		if err != nil {
			return err
		}

		log.Debugf("成功打包 %s ，共写入了 %d 字节的数据\n", fileName, n)

		return nil
	})
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
