package gcloudstorage

import (
	"testing"
)

func TestGCloudStorage_Path(t *testing.T) {

	tests := []struct {
		name           string
		bucket         string
		baseDir        string
		baseURI        string
		image          string
		safeChars      string
		expectedPath   string
		expectedBucket string
		expectedOk     bool
	}{
		{
			name:         "defaults ok",
			image:        "/foo/bar",
			expectedPath: "foo/bar",
			expectedOk:   true,
		},
		{
			name:         "escape unsafe chars",
			image:        "/foo/b{:}ar",
			expectedPath: "foo/b%7B%3A%7Dar",
			expectedOk:   true,
		},
		{
			name:         "escape safe chars",
			image:        "/foo/b{:}ar",
			expectedPath: "foo/b{%3A}ar",
			safeChars:    "{}",
			expectedOk:   true,
		},
		{
			name:         "path under with base uri",
			baseDir:      "home/imagor",
			baseURI:      "/foo",
			image:        "/foo/bar",
			expectedPath: "home/imagor/bar",
			expectedOk:   true,
		},
		{
			name:         "path under no base uri",
			baseDir:      "/home/imagor",
			image:        "/foo/bar",
			expectedPath: "home/imagor/foo/bar",
			expectedOk:   true,
		},
		{
			name:       "path not under",
			baseDir:    "/home/imagor",
			baseURI:    "/foo",
			image:      "/fooo/bar",
			expectedOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts []Option
			if tt.baseURI != "" {
				opts = append(opts, WithPathPrefix(tt.baseURI))
			}
			if tt.baseDir != "" {
				opts = append(opts, WithBaseDir(tt.baseDir))
			}
			opts = append(opts, WithSafeChars(tt.safeChars))

			s := New(nil, tt.bucket, opts...)
			res, ok := s.Path(tt.image)
			if res != tt.expectedPath || ok != tt.expectedOk {
				t.Errorf("= %s,%v want %s,%v", res, ok, tt.expectedPath, tt.expectedOk)
			}
		})
	}
}
