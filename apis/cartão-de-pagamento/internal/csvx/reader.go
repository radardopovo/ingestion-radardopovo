// Radar do Povo ETL - https://radardopovo.com
package csvx

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"unicode"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var headerNormalizer = transform.Chain(
	norm.NFD,
	runes.Remove(runes.In(unicode.Mn)),
	norm.NFC,
)

type Stream struct {
	File       *os.File
	Reader     *csv.Reader
	HeaderMap  map[string]int
	HeaderList []string
}

func Open(path string) (*Stream, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	decoded := transform.NewReader(f, charmap.ISO8859_1.NewDecoder())
	r := csv.NewReader(decoded)
	r.Comma = ';'
	r.FieldsPerRecord = -1
	r.LazyQuotes = true

	headerRow, err := r.Read()
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("falha lendo header CSV: %w", err)
	}

	headerMap := make(map[string]int, len(headerRow))
	normalizedList := make([]string, 0, len(headerRow))
	for i, h := range headerRow {
		nh := NormalizeHeader(h)
		normalizedList = append(normalizedList, nh)
		headerMap[nh] = i
	}

	return &Stream{
		File:       f,
		Reader:     r,
		HeaderMap:  headerMap,
		HeaderList: normalizedList,
	}, nil
}

func (s *Stream) Close() error {
	return s.File.Close()
}

func (s *Stream) Require(headers ...string) error {
	missing := make([]string, 0, len(headers))
	for _, h := range headers {
		if _, ok := s.HeaderMap[NormalizeHeader(h)]; !ok {
			missing = append(missing, h)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("headers obrigatorios ausentes: %s", strings.Join(missing, ", "))
	}
	return nil
}

func (s *Stream) Get(row []string, header string) string {
	idx, ok := s.HeaderMap[NormalizeHeader(header)]
	if !ok || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(row[idx], "\uFEFF"))
}

func NormalizeHeader(v string) string {
	v = strings.TrimPrefix(v, "\uFEFF")
	v = strings.TrimSpace(v)
	v = strings.Join(strings.Fields(v), " ")
	v = strings.ToLower(v)
	if normalized, _, err := transform.String(headerNormalizer, v); err == nil {
		v = normalized
	}
	return v
}
