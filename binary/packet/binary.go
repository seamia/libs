// Copyright 2017-2025 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package packet

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/seamia/libs"
	"github.com/seamia/libs/config"
	"github.com/seamia/libs/iox"
)

const (
	expectedSignature        int32 = 0x77736274 // "wsbt" web socket binary transfer
	binaryPacketPrefixSize         = 4 + 2
	maximumAllowedHeaderSize       = 16 * 1024 // 16K "should be enough for everybody"

	headerBlobSize = "blob.size"

	writeTimeout = time.Second * 30
)

var (
	errSignature = errors.New("signature")
	errTooBig    = errors.New("too big")
)

func CreateBinaryPacket(header libs.Msi, blob []byte, trace libs.Tracer) ([]byte, error) {
	if len(blob) != 0 {
		storedBlobSize := libs.GetInt(header, headerBlobSize, -1)
		if storedBlobSize != len(blob) {
			trace("wrong blob size. %v instead of %v", storedBlobSize, len(blob))
		}
	}

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, expectedSignature)
	if err != nil {
		trace("failed to write signature: %v", err)
		return nil, err
	}

	var (
		headerData []byte
		headerSize int16
	)

	if len(header) > 0 {
		if headerData, err = json.Marshal(header); err != nil {
			trace("failed to marshal header: %v", err)
			return nil, err
		} else if len(headerData) > maximumAllowedHeaderSize {
			trace("header too big: maximum allowed header size exceeded: %v", len(headerData))
			return nil, errTooBig
		}
		headerSize = int16(len(headerData))
	}

	if err := binary.Write(buf, binary.BigEndian, headerSize); err != nil {
		trace("failed to write header size: %v", err)
		return nil, err
	}

	if headerSize > 0 {
		if _, err := buf.Write(headerData); err != nil {
			trace("failed to write header blob: %v", err)
			return nil, err
		}
	}
	if _, err := buf.Write(blob); err != nil {
		trace("failed to write blob: %v", err)
		return nil, err
	}

	return buf.Bytes(), nil
}

func WriteBinary(conn io.Writer, header libs.Msi, blob []byte, trace libs.Tracer) error {
	header[headerBlobSize] = len(blob)

	raw, err := CreateBinaryPacket(header, blob, trace)
	if err != nil {
		trace("failed to create binary packet: %v", err)
		return err
	}

	/*
		if Flag("debug") {
			trace("writing binary blob: %v bytes", len(raw))
			saveBlob(getUniqueName(nil), raw)
		}
	*/

	if config.Flag("write.deadline") {
		if wd, found := conn.(libs.WriteDeadline); found && wd != nil {
			trace("setting write deadline")
			wd.SetWriteDeadline(time.Now().Add(writeTimeout))
		}
	}

	return iox.WriteAll(conn, raw, trace)
}

func ReadBinaryPacket(conn io.Reader, trace libs.Tracer) (*libs.BinaryPacket, error) {
	prefix := make([]byte, binaryPacketPrefixSize)
	size, err := conn.Read(prefix)
	if err != nil {
		return nil, err
	} else if size != binaryPacketPrefixSize {
		return nil, errSignature
	}

	var (
		// IF YOU CHANGE THESE TYPES --> ADJUST binaryPacketPrefixSize value
		signature  int32
		headerSize int16
	)

	buf := bytes.NewReader(prefix)
	if err := binary.Read(buf, binary.BigEndian, &signature); err != nil {
		trace("failed to read signature: %v", err)
		return nil, err
	}
	if signature != expectedSignature {
		trace("invalid signature: %v (%x)", signature, signature)
		return nil, errSignature
	}
	if err := binary.Read(buf, binary.BigEndian, &headerSize); err != nil {
		trace("failed to read headerSize: %v", err)
		return nil, err
	}

	if (headerSize <= 0) || (headerSize > maximumAllowedHeaderSize) {
		trace("invalid headerSize: %v", headerSize)
		return nil, errSignature
	}

	headerBuf := make([]byte, headerSize)
	size, err = conn.Read(headerBuf)
	if err != nil {
		return nil, err
	} else if size != int(headerSize) {
		trace("invalid header size: %v vs %v", size, int(headerSize))
		return nil, errSignature
	}

	var header libs.Msi
	if err := json.Unmarshal(headerBuf, &header); err != nil {
		trace("failed to unmarshal header: %v", err)
		return nil, err
	}

	blobSize := 0
	if value, found := header[headerBlobSize]; found {
		if digit, found := value.(float64); found {
			blobSize = int(digit)
		} else {
			trace("blob.size is not a number; %T", value)
		}
	} else {
		trace("blob.size is not found")
	}

	bp := &libs.BinaryPacket{
		Header: header,
	}

	if blobSize > 0 {
		blob := make([]byte, blobSize)
		if err := iox.ReadAll(conn, blob, trace); err != nil {
			trace("failed to read blob: %v", err)
			return nil, errSignature
		}
		bp.Blob = blob
	}

	return bp, nil
}
