package github_files_provider

import (
	"io"
	"os"
)

type TempFile struct {
	*os.File
	len int
}

func NewTempFile(r io.Reader, maxSizeBytes int) (_ *TempFile, err error) {
	tf := &TempFile{}

	tf.File, err = os.CreateTemp("", "")
	if err != nil {
		return nil, err
	}

	lr := &LimitedReader{Reader: r, Remaining: maxSizeBytes}
	_, err = io.Copy(tf.File, lr)
	if err != nil {
		tf.Close()
		return nil, err
	}

	tf.len = maxSizeBytes - lr.Remaining
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
