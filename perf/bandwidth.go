package perf

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"ulambda/fslib"
)

const (
	MB = 1000000
)

var bucket = "9ps3"
var key = "write-bandwidth-test"
var fname = "name/fs/bigfile.txt"

type BandwidthTest struct {
	mb     int
	memfs  bool
	client *s3.Client
	*fslib.FsLib
}

func MakeBandwidthTest(args []string) (*BandwidthTest, error) {
	if len(args) < 2 {
		return nil, errors.New("MakeBandwidthTest: too few arguments")
	}
	log.Printf("MakeBandwidthTest: %v\n", args)

	t := &BandwidthTest{}
	t.FsLib = fslib.MakeFsLib("write-bandwidth-test")

	mb, err := strconv.Atoi(args[0])
	t.mb = mb
	if err != nil {
		log.Fatalf("Invalid num MB: %v, %v\n", args[0], err)
	}

	if args[1] == "memfs" {
		t.memfs = true
	} else if args[1] == "s3" {
		t.memfs = false
	} else {
		log.Fatalf("Unknown test type: %v", args[1])
	}

	// Set up s3 client
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithSharedConfigProfile("default"))
	if err != nil {
		log.Fatalf("Failed to load SDK configuration %v", err)
	}

	t.client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return t, nil
}

func (t *BandwidthTest) FillBuf(buf []byte) {
	for i := range buf {
		buf[i] = byte(i)
	}
}

func (t *BandwidthTest) S3Write(buf []byte) time.Duration {
	r1 := bytes.NewReader(buf)
	input := &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   r1,
	}
	start := time.Now()
	_, err := t.client.PutObject(context.TODO(), input)
	end := time.Now()
	elapsed := end.Sub(start)
	if err != nil {
		log.Printf("Error putting s3 object: %v", err)
	}
	return elapsed
}

func (t *BandwidthTest) S3Read(buf []byte) time.Duration {
	// setup
	region := "bytes=0-" + strconv.Itoa(len(buf))
	input := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Range:  &region,
	}

	// timing
	start := time.Now()
	result, err := t.client.GetObject(context.TODO(), input)
	end := time.Now()
	elapsed := end.Sub(start)
	if err != nil {
		log.Fatalf("Error getting s3 object: %v", err)
	}
	buf2 := make([]byte, len(buf))
	n := 0
	for {
		n1, err := result.Body.Read(buf2[n:])
		n += n1
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error reading s3 object result: %v", err)
		}
	}
	if n != len(buf) {
		log.Fatalf("Length of s3 read buffer didn't match: %v, %v", n, len(buf))
	}
	for i := range buf2 {
		if buf2[i] != buf[i] {
			log.Fatalf("S3 Read buf didn't match written buf at index %v", i)
		}
	}
	return elapsed
}

func (t *BandwidthTest) MemfsWrite(buf []byte) time.Duration {
	// setup
	err := t.MakeFile(fname, []byte{})
	if err != nil && err.Error() != "Name exists" {
		log.Fatalf("Error creating file: %v", err)
	}

	// timing
	start := time.Now()
	err = t.WriteFile(fname, buf)
	end := time.Now()
	elapsed := end.Sub(start)

	return elapsed
}

func (t *BandwidthTest) MemfsRead(buf []byte) time.Duration {
	// timing
	start := time.Now()
	buf2, err := t.ReadFile(fname)
	end := time.Now()
	elapsed := end.Sub(start)

	for i := range buf2 {
		if buf2[i] != buf[i] {
			log.Fatalf("Memfs Read buf didn't match written buf at index %v", i)
		}
	}

	// cleanup
	err = t.Remove(fname)
	if err != nil {
		log.Printf("Error removing file: %v", err)
	}
	return elapsed
}

func (t *BandwidthTest) Work() {
	buf := make([]byte, t.mb*MB)
	t.FillBuf(buf)
	var elapsedWrite time.Duration
	var elapsedRead time.Duration
	if t.memfs {
		elapsedWrite = t.MemfsWrite(buf)
		elapsedRead = t.MemfsRead(buf)
	} else {
		elapsedWrite = t.S3Write(buf)
		elapsedRead = t.S3Read(buf)
	}
	log.Printf("Bytes: %v", t.mb*MB)
	log.Printf("Write Time: %v (usec)", elapsedWrite.Microseconds())
	log.Printf("Write Throughput: %f (MB/sec)", float64(t.mb)/elapsedWrite.Seconds())
	log.Printf("Read Time: %v (usec)", elapsedRead.Microseconds())
	log.Printf("Read Throughput: %f (MB/sec)", float64(t.mb)/elapsedRead.Seconds())
}
