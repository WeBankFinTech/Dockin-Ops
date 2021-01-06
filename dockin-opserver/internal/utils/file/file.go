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

package file

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/webankfintech/dockin-opserver/internal/log"

	"github.com/pkg/errors"
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
	return strings.TrimLeft(file, "/")
}

func MakeTar(srcPath, destPath string, writer io.Writer) error {
	tarWriter := tar.NewWriter(writer)
	defer tarWriter.Close()

	srcPath = path.Clean(srcPath)
	destPath = path.Clean(destPath)
	return recursiveTar(path.Dir(srcPath), path.Base(srcPath), path.Dir(destPath), path.Base(destPath), tarWriter)
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
	return relative == "." || relative == StripPathShortcuts(relative)
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

type fileSpec struct {
	PodNamespace string
	PodName      string
	File         string
}

var (
	errFileSpecDoesntMatchFormat = errors.New("Filespec must match the canonical format: [[namespace/]pod:]file/path")
	errFileCannotBeEmpty         = errors.New("Filepath can not be empty")
)

func extractFileSpec(arg string) (fileSpec, error) {
	if i := strings.Index(arg, ":"); i == -1 {
		return fileSpec{File: arg}, nil
	} else if i > 0 {
		file := arg[i+1:]
		pod := arg[:i]
		pieces := strings.Split(pod, "/")
		if len(pieces) == 1 {
			return fileSpec{
				PodName: pieces[0],
				File:    file,
			}, nil
		}
		if len(pieces) == 2 {
			return fileSpec{
				PodNamespace: pieces[0],
				PodName:      pieces[1],
				File:         file,
			}, nil
		}
	}

	return fileSpec{}, errFileSpecDoesntMatchFormat
}

func UntarFile(reader io.Reader, dest string) error {
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Logger.Warnf("ERROR: cannot read tar file, error=[%v]\n", err)
			return err
		}

		_, err = json.Marshal(header)
		if err != nil {
			log.Logger.Warnf("ERROR: cannot parse header, error=[%v]\n", err)
			return err
		}

		info := header.FileInfo()
		if info.IsDir() {
			filePath := filepath.Join(dest, header.Name)
			if err = os.MkdirAll(filePath, 0755); err != nil {
				log.Logger.Warnf("ERROR: cannot mkdir file, error=[%v]\n", err)
				return err
			}
		} else {
			filePath := filepath.Join(dest, path.Dir(header.Name))
			if err = os.MkdirAll(filePath, 0755); err != nil {
				log.Logger.Warnf("ERROR: cannot file mkdir file, error=[%v]\n", err)
				return err
			}

			filePath = filepath.Join(dest, header.Name)
			file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
			if err != nil {
				log.Logger.Warnf("ERROR: cannot open file, error=[%v]\n", err)
				return err
			}
			defer file.Close()

			_, err = io.Copy(file, tarReader)
			if err != nil {
				log.Logger.Warnf("ERROR: cannot write file, error=[%v]\n", err)
				return err
			}
		}
	}
	return nil
}
