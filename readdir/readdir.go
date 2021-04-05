package readdir

import (
	"os"
	"path/filepath"
	"regexp"
)

type Reader struct {
	dirname string
	pattern string
	re      *regexp.Regexp
}

func New(dirname, pattern string, re *regexp.Regexp) *Reader {
	return &Reader{
		dirname: filepath.Clean(dirname),
		pattern: pattern,
		re:      re,
	}
}

func (r *Reader) String() string {
	return r.dirname + "/" + r.pattern
}

func (r *Reader) MatchString(s string) bool {
	if r.re == nil {
		return r.pattern == s
	}
	return r.re.MatchString(s)
}

func (r *Reader) Read() ([]string, error) {
	f, err := os.Open(r.dirname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	entries, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	list := make([]string, 0, len(entries)/2)
	for _, entry := range entries {
		if r.MatchString(entry) {
			list = append(list, filepath.Join(r.dirname, entry))
		}
	}

	return list, nil
}
