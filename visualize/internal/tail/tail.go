// Package tail follows an EmbeddedApp.py log file and yields its records.
package tail

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"time"
)

// EmbeddedApp.py:17 writes "%(asctime)s %(levelname)s [%(event)s] %(message)s",
// so a record splits positionally into date, time, level, bracketed event and
// message. The message carries its own timestamps and JSON, so it is never
// scanned for a delimiter.
const (
	fields = 5
	layout = "2006-01-02 15:04:05.000"
)

const (
	poll      = 250 * time.Millisecond
	seekBlock = 64 * 1024
)

var errRotated = errors.New("log file rotated")

// Line is one parsed record, as the UI consumes it.
type Line struct {
	Ts    string `json:"ts"`
	Epoch int64  `json:"epoch"`
	Level string `json:"level"`
	Type  string `json:"type"`
	Msg   string `json:"msg"`
}

// Tail opens path, rewinds seed records from the end and calls fn for every
// record from there on, following the file across rotations. It runs on the
// caller's goroutine and returns when ctx is done or fn reports an error.
func Tail(ctx context.Context, path string, seed int, fn func(Line) error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { f.Close() }()

	if err := rewind(f, seed); err != nil {
		return err
	}
	for {
		err := follow(ctx, f, path, fn)
		if !errors.Is(err, errRotated) {
			return err
		}
		// The writer rolled over, so the replacement is read from the top.
		f.Close()
		if f, err = os.Open(path); err != nil {
			return err
		}
	}
}

func follow(ctx context.Context, f *os.File, path string, fn func(Line) error) error {
	reader := bufio.NewReader(f)
	var pending strings.Builder
	for {
		for {
			chunk, err := reader.ReadString('\n')
			pending.WriteString(chunk)
			if err != nil {
				break // no terminator yet, so the writer is mid-record
			}
			record := pending.String()
			pending.Reset()
			if err := fn(split(record)); err != nil {
				return err
			}
		}

		// Checked only once the open file is drained, so records written just
		// before the rename are delivered before the handle is let go.
		if rotated(f, path) {
			return errRotated
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(poll):
		}
	}
}

func rotated(f *os.File, path string) bool {
	// Between the rename and the create, path briefly names nothing. Keeping
	// the open handle lets the next poll find the replacement.
	onDisk, err := os.Stat(path)
	if err != nil {
		return false
	}
	open, err := f.Stat()
	if err != nil {
		return false
	}
	return !os.SameFile(onDisk, open)
}

// rewind positions f at the start of the last n records, or at the top if the
// file holds fewer.
func rewind(f *os.File, n int) error {
	end, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	buf := make([]byte, seekBlock)
	count := 0
	pos := end
	for pos > 0 {
		step := int64(seekBlock)
		if pos < step {
			step = pos
		}
		pos -= step
		if _, err := f.ReadAt(buf[:step], pos); err != nil {
			return err
		}
		for i := int(step) - 1; i >= 0; i-- {
			if buf[i] != '\n' {
				continue
			}
			if pos+int64(i) == end-1 {
				continue // the last record's own terminator is not a boundary
			}
			count++
			if count == n {
				_, err := f.Seek(pos+int64(i)+1, io.SeekStart)
				return err
			}
		}
	}

	_, err = f.Seek(0, io.SeekStart)
	return err
}

// split parses one record. Anything that does not match LOG_FORMAT comes back
// whole in Msg.
func split(raw string) Line {
	text := strings.TrimRight(raw, "\r\n")
	part := strings.SplitN(text, " ", fields)
	if len(part) < fields-1 {
		return Line{Msg: text}
	}

	event := part[3]
	if len(event) < 2 || event[0] != '[' || event[len(event)-1] != ']' {
		return Line{Msg: text}
	}
	at, err := time.Parse(layout, part[0]+" "+strings.Replace(part[1], ",", ".", 1))
	if err != nil {
		return Line{Msg: text}
	}

	var msg string
	if len(part) == fields {
		msg = part[4]
	}
	return Line{
		Ts:    part[0] + " " + part[1],
		Epoch: at.UnixMilli(),
		Level: part[2],
		Type:  event[1 : len(event)-1],
		Msg:   msg,
	}
}
