package blog

import (
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"io"
	"os"
)

func Move(f1, f2 string) {
	if err := os.Rename(f1, f2); err == nil {
		return
	}
	if fo, err := os.Create(f2); err != nil {
		fmt.Printf("Move failed (mv %s %s)\n", f1, f2)
		return
	} else {
		if fi, err := os.Open(f1); err != nil {
			fmt.Printf("Move: Can't open file(%s)\n", f1)
			return
		} else {
			io.Copy(fo, fi)
			fi.Close()
		}
		fo.Close()
	}
	os.Remove(f1)
}

func IsExist(file string) bool {
	_, err := os.Stat(file)
	return err == nil || os.IsExist(err)
}

func IsDir(file string) bool {
	f, err := os.Stat(file)
	if err == nil && f.IsDir() {
		return true
	}
	return false
}

func Md5(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func Gz(fin string, fout string) error {
	fo, err := os.Create(fout)
	if err != nil {
		return err
	}
	defer fo.Close()
	gw := gzip.NewWriter(fo)
	defer gw.Close()
	fi, err := os.Open(fin)
	if err != nil {
		return err
	}
	defer fi.Close()
	io.Copy(gw, fi)
	return err
}

func Ungz(fin string, fout string) error {
	fo, err := os.Create(fout)
	if err != nil {
		return err
	}
	defer fo.Close()
	fi, err := os.Open(fin)
	if err != nil {
		return err
	}
	defer fi.Close()
	gw, _ := gzip.NewReader(fi)
	defer gw.Close()
	io.Copy(fo, gw)
	return err
}
