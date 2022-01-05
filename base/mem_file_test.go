package base

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"testing"
)

// DRY Principle
func CalHexMd5WithBufSize(t *testing.T, bufSize, fileSize int64) string {
	f := NewMemFile(
		WithDefaultContent(),
		WithFileSize(fileSize),
	)

	buf := make([]byte, bufSize)
	md5x := md5.New()
	for {
		readed, err := f.Read(buf)
		if err == io.EOF {
			break
		}
		md5x.Write(buf[:readed])
	}

	si, _ := f.Stat()
	t.Logf("file readOffset:%d", si.(*MemFileInfo).ReadOffset())
	return hex.EncodeToString(md5x.Sum(nil))
}

func TestMemFile(t *testing.T) {
	md51Hex := CalHexMd5WithBufSize(t, 128, 1<<20)
	md52Hex := CalHexMd5WithBufSize(t, 256, 1<<20)
	md53Hex := CalHexMd5WithBufSize(t, 37, 1<<20)
	if md51Hex != md52Hex || md52Hex != md53Hex || md53Hex != md51Hex {
		t.Errorf("checksum is not equal, md5_1:%s md5_2:%s md5_3:%s", md51Hex, md52Hex, md53Hex)
	} else {
		t.Logf("checksum is equal, md5_1:%s md5_2:%s md5_3:%s, check succ!!", md51Hex, md52Hex, md53Hex)
	}

	md54Hex := CalHexMd5WithBufSize(t, 128, 512<<20+23471)
	md55Hex := CalHexMd5WithBufSize(t, 256, 512<<20+23471)
	if md54Hex != md55Hex {
		t.Errorf("checksum is not equal, md5_4:%s md5_5:%s", md54Hex, md55Hex)
	} else {
		t.Logf("checksum is equal, md5_4:%s md5_:%s, check succ!!", md54Hex, md55Hex)
	}
}
