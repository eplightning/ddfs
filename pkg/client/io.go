package client

import "io"

type IndexFillReader struct {
	Fill byte
	Size int64
	cur  int64
}

func (f *IndexFillReader) Read(b []byte) (n int, err error) {
	max := len(b)
	remaining := int(f.Size - f.cur)
	if remaining <= 0 {
		return 0, io.EOF
	}

	if max > remaining {
		max = remaining
	}

	f.cur += int64(max)
	if f.cur >= f.Size {
		err = io.EOF
	}

	for i := 0; i < max; i++ {
		b[i] = f.Fill
	}

	n = max
	return
}

func (f *IndexFillReader) Seek(offset int64, whence int) (int64, error) {
	if whence == io.SeekCurrent {
		f.cur += offset
	} else if whence == io.SeekStart {
		f.cur = offset
	} else {
		f.cur = f.Size + offset
	}

	return f.cur, nil
}
