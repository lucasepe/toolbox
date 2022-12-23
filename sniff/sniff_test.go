// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sniff

import (
	"testing"
)

var sniffTests = []struct {
	desc        string
	data        []byte
	contentType string
}{
	// Some nonsense.
	{"Empty", []byte{}, "text/plain; charset=utf-8"},
	{"Binary", []byte{1, 2, 3}, "application/octet-stream"},

	{"HTML document #1", []byte(`<HtMl><bOdY>blah blah blah</body></html>`), "text/html; charset=utf-8"},
	{"HTML document #2", []byte(`<HTML></HTML>`), "text/html; charset=utf-8"},
	{"HTML document #3 (leading whitespace)", []byte(`   <!DOCTYPE HTML>...`), "text/html; charset=utf-8"},
	{"HTML document #4 (leading CRLF)", []byte("\r\n<html>..."), "text/html; charset=utf-8"},

	{"Plain text", []byte(`This is not HTML. It has ☃ though.`), "text/plain; charset=utf-8"},

	{"XML", []byte("\n<?xml!"), "text/xml; charset=utf-8"},

	// Image types.
	{"Windows icon", []byte("\x00\x00\x01\x00"), "image/x-icon"},
	{"Windows cursor", []byte("\x00\x00\x02\x00"), "image/x-icon"},
	{"BMP image", []byte("BM..."), "image/bmp"},
	{"GIF 87a", []byte(`GIF87a`), "image/gif"},
	{"GIF 89a", []byte(`GIF89a...`), "image/gif"},
	{"WEBP image", []byte("RIFF\x00\x00\x00\x00WEBPVP"), "image/webp"},
	{"PNG image", []byte("\x89PNG\x0D\x0A\x1A\x0A"), "image/png"},
	{"JPEG image", []byte("\xFF\xD8\xFF"), "image/jpeg"},

	// Audio types.
	{"MIDI audio", []byte("MThd\x00\x00\x00\x06\x00\x01"), "audio/midi"},
	{"MP3 audio/MPEG audio", []byte("ID3\x03\x00\x00\x00\x00\x0f"), "audio/mpeg"},
	{"WAV audio #1", []byte("RIFFb\xb8\x00\x00WAVEfmt \x12\x00\x00\x00\x06"), "audio/wave"},
	{"WAV audio #2", []byte("RIFF,\x00\x00\x00WAVEfmt \x12\x00\x00\x00\x06"), "audio/wave"},
	{"AIFF audio #1", []byte("FORM\x00\x00\x00\x00AIFFCOMM\x00\x00\x00\x12\x00\x01\x00\x00\x57\x55\x00\x10\x40\x0d\xf3\x34"), "audio/aiff"},

	{"OGG audio", []byte("OggS\x00\x02\x00\x00\x00\x00\x00\x00\x00\x00\x7e\x46\x00\x00\x00\x00\x00\x00\x1f\xf6\xb4\xfc\x01\x1e\x01\x76\x6f\x72"), "application/ogg"},
	{"Must not match OGG", []byte("owow\x00"), "application/octet-stream"},
	{"Must not match OGG", []byte("oooS\x00"), "application/octet-stream"},
	{"Must not match OGG", []byte("oggS\x00"), "application/octet-stream"},

	// Video types.
	{"MP4 video", []byte("\x00\x00\x00\x18ftypmp42\x00\x00\x00\x00mp42isom<\x06t\xbfmdat"), "video/mp4"},
	{"AVI video #1", []byte("RIFF,O\n\x00AVI LISTÀ"), "video/avi"},
	{"AVI video #2", []byte("RIFF,\n\x00\x00AVI LISTÀ"), "video/avi"},

	// Font types.
	// {"MS.FontObject", []byte("\x00\x00")},
	{"TTF sample  I", []byte("\x00\x01\x00\x00\x00\x17\x01\x00\x00\x04\x01\x60\x4f"), "font/ttf"},
	{"TTF sample II", []byte("\x00\x01\x00\x00\x00\x0e\x00\x80\x00\x03\x00\x60\x46"), "font/ttf"},

	{"OTTO sample  I", []byte("\x4f\x54\x54\x4f\x00\x0e\x00\x80\x00\x03\x00\x60\x42\x41\x53\x45"), "font/otf"},

	{"woff sample  I", []byte("\x77\x4f\x46\x46\x00\x01\x00\x00\x00\x00\x30\x54\x00\x0d\x00\x00"), "font/woff"},
	{"woff2 sample", []byte("\x77\x4f\x46\x32\x00\x01\x00\x00\x00"), "font/woff2"},
	{"wasm sample", []byte("\x00\x61\x73\x6d\x01\x00"), "application/wasm"},

	// Archive types
	{"RAR v1.5-v4.0", []byte("Rar!\x1A\x07\x00"), "application/x-rar-compressed"},
	{"RAR v5+", []byte("Rar!\x1A\x07\x01\x00"), "application/x-rar-compressed"},
	{"Incorrect RAR v1.5-v4.0", []byte("Rar \x1A\x07\x00"), "application/octet-stream"},
	{"Incorrect RAR v5+", []byte("Rar \x1A\x07\x01\x00"), "application/octet-stream"},
}

func TestDetectContentType(t *testing.T) {
	for _, tt := range sniffTests {
		ct := DetectContentType(tt.data)
		if ct != tt.contentType {
			t.Errorf("%v: DetectContentType = %q, want %q", tt.desc, ct, tt.contentType)
		}
	}
}
