package main

import (
	"io"
	"sync/atomic"
)

type MeteredReader struct {
	R         io.Reader
	ReadCount int64
}

func (mRdr *MeteredReader) Read(buf []byte) (int, error) {
	n, err := mRdr.R.Read(buf)
	atomic.AddInt64(&mRdr.ReadCount, int64(n))
	return n, err
}

type MeteredWriter struct {
	W          io.Writer
	WriteCount int64
}

func (mWtr *MeteredWriter) Write(buf []byte) (int, error) {
	n, err := mWtr.W.Write(buf)
	atomic.AddInt64(&mWtr.WriteCount, int64(n))
	return n, err
}
