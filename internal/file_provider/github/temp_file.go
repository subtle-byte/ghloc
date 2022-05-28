package github

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

type TempFile struct {
	*os.File
	len int
}

func NewTempFile(r io.Reader) (_ *TempFile, err error) {
	tf := &TempFile{}

	tf.File, err = ioutil.TempFile("", "")
	if err != nil {
		return nil, err
	}
	log.Print("temp file: ", tf.File.Name())

	lr := &LimitedReader{Reader: r, Remaining: maxZipSize}
	_, err = io.Copy(tf.File, lr)
	if err != nil {
		tf.Close()
		return nil, err
	}

	tf.len = maxZipSize - lr.Remaining
	return tf, nil
}

func (tf *TempFile) Close() error {
	err1 := tf.File.Close()
	err2 := os.Remove(tf.File.Name())
	if err1 != nil {
		return err1
	}
	return err2
}

func (tf *TempFile) Len() int {
	return tf.len
}
