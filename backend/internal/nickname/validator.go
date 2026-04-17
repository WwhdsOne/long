package nickname

import (
	"bufio"
	"embed"
	"errors"
	"io/fs"
	"slices"
	"strings"
)

var ErrSensitiveNickname = errors.New("sensitive nickname")

//go:embed lexicon/upstream
var embeddedLexiconFS embed.FS

// Validator checks whether a nickname contains blocked terms.
type Validator struct {
	terms []string
}

// NewValidator builds a validator from an explicit word list.
func NewValidator(terms []string) *Validator {
	normalized := make([]string, 0, len(terms))
	seen := make(map[string]struct{}, len(terms))

	for _, term := range terms {
		cleaned := strings.ToLower(strings.TrimSpace(term))
		if cleaned == "" {
			continue
		}
		if _, exists := seen[cleaned]; exists {
			continue
		}

		seen[cleaned] = struct{}{}
		normalized = append(normalized, cleaned)
	}

	// Longer terms first reduces accidental short-word matches taking precedence.
	slices.SortFunc(normalized, func(left, right string) int {
		if len(left) == len(right) {
			return strings.Compare(left, right)
		}
		return len(right) - len(left)
	})

	return &Validator{terms: normalized}
}

// NewSensitiveLexiconValidator returns a validator backed by every vendored
// text lexicon from konsheng/Sensitive-lexicon.
func NewSensitiveLexiconValidator() *Validator {
	terms := make([]string, 0, 2048)

	_ = fs.WalkDir(embeddedLexiconFS, "lexicon/upstream", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txt") {
			return err
		}

		content, readErr := embeddedLexiconFS.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		scanner := bufio.NewScanner(strings.NewReader(string(content)))
		for scanner.Scan() {
			terms = append(terms, scanner.Text())
		}

		return scanner.Err()
	})

	return NewValidator(terms)
}

// Validate returns ErrSensitiveNickname when the nickname contains a blocked term.
func (v *Validator) Validate(value string) error {
	if v == nil {
		return nil
	}

	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return nil
	}

	for _, term := range v.terms {
		if strings.Contains(normalized, term) {
			return ErrSensitiveNickname
		}
	}

	return nil
}
