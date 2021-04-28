package readdir

import (
	"os"
	"path/filepath"
	"regexp"
)

type Reader struct {
	dirname  string
	filename string
	re       *regexp.Regexp
}

type FilenameAs int

const (
	AsPlain FilenameAs = 0x0
	AsRegex FilenameAs = 0x1
	AsPosix FilenameAs = 0x2
)

func New(dirname, filename string, asa FilenameAs) (*Reader, error) {
	var re *regexp.Regexp
	var err error
	if asa&AsPosix != 0 {
		re, err = regexp.CompilePOSIX(filename)
	} else if asa&AsRegex != 0 {
		re, err = regexp.Compile(filename)
	}
	if err != nil {
		return nil, err
	}

	return &Reader{
		dirname:  filepath.Clean(dirname),
		filename: filename,
		re:       re,
	}, nil
}

func (r *Reader) String() string {
	return r.dirname + "/" + r.filename
}

func (r *Reader) MatchString(s string) bool {
	if r.re == nil {
		return r.filename == s
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
