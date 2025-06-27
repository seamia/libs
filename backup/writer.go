package backup

import (
	"encoding/json"
	"fmt"
	"io"
)

type writeBuffer struct {
	media    io.Writer
	buffer   []interface{}
	capacity int
}

func createBuffer(media io.Writer, batch int) (*writeBuffer, error) {
	if media == nil || batch < 1 {
		return nil, errNilAccess
	}
	buffer := &writeBuffer{
		media:    media,
		buffer:   nil,
		capacity: batch,
	}
	return buffer, success
}

func (wb *writeBuffer) Add(what interface{}) error {
	if wb == nil || what == nil {
		return errNilAccess
	}

	if wb.buffer == nil {
		wb.buffer = make([]interface{}, 0, wb.capacity)
	}

	wb.buffer = append(wb.buffer, what)

	if len(wb.buffer) == wb.capacity {
		return wb.Flush()
	}
	return success
}

func (wb *writeBuffer) Flush() error {
	if len(wb.buffer) > 0 {
		data, err := json.MarshalIndent(wb.buffer, "", "    ")
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintf(wb.media, "\n%v\n.\n", len(data)); err != nil {
			return err
		}

		if _, err := wb.media.Write(data); err != nil {
			return err
		}
		wb.buffer = nil
	}
	return success
}
