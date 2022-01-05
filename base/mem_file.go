/*
 *	This tool component is mainly to facilitate testing,
 *	using a small amount of memory to simulate a large file
 */
package base

import (
	"errors"
	"io"
	"io/fs"
	"time"
)

const (
	Ctt = `An example of this pattern is an ingestion engine service built on 
Windows Azure to provide near real-time Facebook and Twitter 
search. This service is one part of a larger data processing 
pipeline that provides publically searchable content (via our 
search engine, Bing) within 15 seconds of a Facebook or Twitter 
user’s posting or status update. Facebook and Twitter send the 
raw public content to WAS (e.g., user postings, user status 
updates, etc.) to be made publically searchable. This content is 
stored in WAS Blobs. The ingestion engine annotates this data 
with user auth, spam, and adult scores; content classification; and 
classification for language and named entities. In addition, the 
engine crawls and expands the links in the data. While 
processing, the ingestion engine accesses WAS Tables at high 
rates and stores the results back into Blobs. These Blobs are then 
folded into the Bing search engine to make the content publically 
searchable. The ingestion engine uses Queues to manage the flow 
of work, the indexing jobs, and the timing of folding the results 
into the search engine. As of this writing, the ingestion engine for 
Facebook and Twitter keeps around 350TB of data in WAS 
(before replication). In terms of transactions, the ingestion engine 
has a peak traffic load of around 40,000 transactions per second 
and does between two to three billion transactions per day (see 
Section 7 for discussion of additional workload profiles)`
)

type MemFileInfo struct {
	name       string
	size       int64
	readOffset int64
}

func (mfi *MemFileInfo) Name() string {
	return mfi.name
}

func (mfi *MemFileInfo) Size() int64 {
	return mfi.size
}

func (mfi *MemFileInfo) Mode() fs.FileMode {
	return fs.ModePerm
}

func (mfi *MemFileInfo) ModTime() time.Time {
	return time.Now()
}

func (mfi *MemFileInfo) IsDir() bool {
	return false
}

func (mfi *MemFileInfo) Sys() interface{} {
	return nil
}

func (mfi *MemFileInfo) ReadOffset() int64 {
	return mfi.readOffset
}

type MemFile struct {
	ctt []byte
	MemFileInfo
}

func (f *MemFile) Stat() (fs.FileInfo, error) {
	return &(f.MemFileInfo), nil
}

func (f *MemFile) Read(buf []byte) (int, error) {
	copyOffset := 0

	for {
		if f.readOffset >= f.size {
			if copyOffset > 0 {
				break
			}
			return 0, io.EOF
		}
		readOffset := int(f.readOffset % (int64(len(f.ctt))))
		copyLen := MinInt(len(buf[copyOffset:]), len(f.ctt[readOffset:]))
		copyLen = MinInt(copyLen, int(f.size-f.readOffset))
		if copyLen == 0 {
			return copyOffset, nil
		}

		copy(buf[copyOffset:copyOffset+copyLen], f.ctt[readOffset:])
		copyOffset += copyLen
		f.readOffset += int64(copyLen)
	}

	return copyOffset, nil
}

func (f *MemFile) Close() error {
	if f.readOffset == -1 {
		return fs.ErrClosed
	}
	f.readOffset = -1
	return nil
}

func (f *MemFile) Seek(offset int64) error {
	if offset < 0 {
		return errors.New("offset cannot be negative")
	}

	if f.size <= offset {
		return io.EOF
	}

	f.size = offset
	return nil
}

type memFileOption func(*MemFile)

func WithContent(ctt []byte) memFileOption {
	return func(f *MemFile) {
		f.ctt = ctt
	}
}

func WithFileSize(size int64) memFileOption {
	return func(f *MemFile) {
		if size <= 0 {
			panic("size cannot be negative")
		}
		f.size = size
	}
}

func WithDefaultContent() memFileOption {
	return func(f *MemFile) {
		f.ctt = StringToBytes(Ctt)
	}
}

func NewMemFile(options ...memFileOption) fs.File {
	file := &MemFile{}
	for _, option := range options {
		option(file)
	}
	return file
}
