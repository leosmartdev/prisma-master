package tmsg

import (
	. "prisma/tms"
	"prisma/tms/log"

	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"math/rand"
	"unsafe"
)

const (
	MsgStartMagicNumber     = 0x5453494d53475632
	MsgStartZLibMagicNumber = 0x5453495a4c425632
	MaxMessageSize          = 32 * 1024 * 1024 // 32 MB
)

var (
	NetworkByteOrder = binary.BigEndian
	HdrSize          = int(unsafe.Sizeof(TsiMessageHeader{}))
	FtrSize          = int(unsafe.Sizeof(TsiMessageFooter{}))

	RetryError = errors.New("Recoverable error, retry")
)

type TsiMessageHeader struct {
	MsgStart    uint64
	MsgRandomId uint64
	MsgSize     uint64
}

type TsiMessageFooter struct {
	MsgRandomId   uint64
	MsgSha512Hash [64]byte
}

func ReadTsiMessage(ctxt context.Context, r *bufio.Reader) (*TsiMessage, error) {
	msg, _, err := ReadTsiMessageExtended(ctxt, r)
	return msg, err
}

func ReadTsiMessageExtended(ctxt context.Context, r *bufio.Reader) (*TsiMessage, uint64, error) {
	headerBytes, err := r.Peek(HdrSize)
	if err != nil {
		log.Error("Error peeking into header: %v", err)
		return nil, 0, err
	}

	var header TsiMessageHeader
	err = binary.Read(bytes.NewReader(headerBytes), NetworkByteOrder, &header)
	if err != nil {
		log.Error("Error reading header: %v", err)
		return nil, 0, err
	}

	zlibDecompress := false
	switch header.MsgStart {
	case MsgStartMagicNumber:
		// Nothing special
	case MsgStartZLibMagicNumber:
		// Nothing special
		zlibDecompress = true
	default:
		log.Warn("Invalid header magic number: %v", header)
		// Discard 1 byte to advance
		r.Discard(1)
		return nil, header.MsgRandomId, RetryError
	}

	wholeSize := HdrSize + FtrSize + int(header.MsgSize)
	if wholeSize > MaxMessageSize {
		log.Error("Message is larger than maximum size, dropping")
		_, err := r.Discard(wholeSize)
		if err != nil {
			log.Error("Error discarding bytes: %v", err)
			return nil, header.MsgRandomId, err
		}
		return nil, header.MsgRandomId, RetryError
	}

	wholeMsg, err := r.Peek(wholeSize)
	if err != nil {
		log.Error("Error peeking into whole message: %v", err)
		return nil, header.MsgRandomId, err
	}

	footerStart := HdrSize + int(header.MsgSize)
	var footer TsiMessageFooter
	err = binary.Read(bytes.NewReader(wholeMsg[footerStart:]), NetworkByteOrder, &footer)
	if err != nil {
		return nil, header.MsgRandomId, err
	}

	if footer.MsgRandomId != header.MsgRandomId {
		log.Warn("Footer random id doesn't match: %v, %v", header, footer)
		r.Discard(wholeSize) // Discard invalid message
		return nil, header.MsgRandomId, RetryError
	}

	headerAndMsg := wholeMsg[:footerStart]
	hash := sha512.Sum512(headerAndMsg)
	if hash != footer.MsgSha512Hash {
		log.Warn("Footer hash mismatch: %v, %v", hash, footer)
		r.Discard(wholeSize) // Discard invalid message
		return nil, header.MsgRandomId, RetryError
	}

	protoMsgOnly := wholeMsg[HdrSize:footerStart]

	if zlibDecompress {
		buff := bytes.NewBuffer(protoMsgOnly)
		r, err := zlib.NewReader(buff)
		if err != nil {
			log.Warn("Error creating zlib decompressor: %v", err)
			return nil, header.MsgRandomId, err
		}
		into := bytes.NewBuffer(nil)
		_, err = into.ReadFrom(r)
		if err != nil {
			log.Warn("Error reading from decompressor: %v", err)
			return nil, header.MsgRandomId, err
		}
		r.Close()
		protoMsgOnly = into.Bytes()
	}

	var msg TsiMessage
	err = proto.Unmarshal(protoMsgOnly, &msg)
	if err != nil {
		log.Warn("Error unmarshalling message: %v", err)
		r.Discard(wholeSize) // Discard invalid message
		return nil, header.MsgRandomId, RetryError
	}

	r.Discard(wholeSize)
	return &msg, header.MsgRandomId, nil
}

func WriteTsiMessage(ctxt context.Context, w *bufio.Writer, msg *TsiMessage) error {
	return WriteTsiMessageExtended(ctxt, w, msg, Opts{ID: uint64(rand.Int63())})
}

type Opts struct {
	ID       uint64
	Compress bool
}

func WriteTsiMessageExtended(ctxt context.Context, w *bufio.Writer, msg *TsiMessage, opts Opts) error {
	pbBytes, err := proto.Marshal(msg)
	if err != nil {
		log.Error("Error marshalling message: %v", err)
		return RetryError
	}

	header := TsiMessageHeader{
		MsgStart:    MsgStartMagicNumber,
		MsgRandomId: opts.ID,
		MsgSize:     uint64(len(pbBytes)),
	}

	if opts.Compress {
		cbuffer := bytes.NewBuffer(nil)
		w, err := zlib.NewWriterLevel(cbuffer, zlib.BestCompression)
		if err != nil {
			log.Error("Error creating compressor: %v", err)
			return err
		}
		w.Write(pbBytes)
		w.Close()
		compressedSize := cbuffer.Len()
		if compressedSize < len(pbBytes) {
			// Only replace if compressed message is smaller
			log.TraceMsg("Compressed message by %f%%", 100.0*(1.0-float32(compressedSize)/float32(len(pbBytes))))

			header.MsgStart = MsgStartZLibMagicNumber
			pbBytes = cbuffer.Bytes()
			header.MsgSize = uint64(len(pbBytes))
		}
	}

	buffer := &bytes.Buffer{}
	err = binary.Write(buffer, NetworkByteOrder, header)
	if err != nil {
		log.Error("Error writing header to buffer: %v", err)
		return RetryError
	}

	buffer.Write(pbBytes)
	footer := TsiMessageFooter{
		MsgRandomId:   header.MsgRandomId,
		MsgSha512Hash: sha512.Sum512(buffer.Bytes()),
	}
	err = binary.Write(buffer, NetworkByteOrder, footer)
	if err != nil {
		log.Error("Error writing footer to buffer: %v", err)
		return RetryError
	}

	if buffer.Len() > MaxMessageSize {
		return errors.New(fmt.Sprintf("Message is too large (%v)", buffer.Len()))
	}

	_, err = buffer.WriteTo(w)
	if err != nil {
		return err
	}
	return w.Flush()
}
