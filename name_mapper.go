package stanza

import "unicode"

func init() {
	SnackCase = snackCase
	TitleUnderscore = titleUnderscore
}

// snackCase converts "CamelCase" → "camel_case".
func snackCase(s string) string {
	if s == "" {
		return s
	}
	var buf []byte
	runes := []rune(s)
	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 {
				prev := runes[i-1]
				if unicode.IsLower(prev) || unicode.IsDigit(prev) {
					buf = append(buf, '_')
				} else if unicode.IsUpper(prev) && i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
					buf = append(buf, '_')
				}
			}
			buf = append(buf, byte(unicode.ToLower(r)))
		} else {
			buf = append(buf, byte(r))
		}
	}
	return string(buf)
}

// titleUnderscore converts "CamelCase" → "Camel_Case".
func titleUnderscore(s string) string {
	if s == "" {
		return s
	}
	var buf []byte
	runes := []rune(s)
	for i, r := range runes {
		if unicode.IsUpper(r) && i > 0 {
			prev := runes[i-1]
			if unicode.IsLower(prev) || unicode.IsDigit(prev) {
				buf = append(buf, '_')
			} else if unicode.IsUpper(prev) && i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
				buf = append(buf, '_')
			}
		}
		buf = append(buf, byte(r))
	}
	return string(buf)
}
