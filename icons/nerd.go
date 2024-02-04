package icons

type nerd struct {
	m map[string]rune
}

func (i *nerd) Lang(val string) string {
	if r, ok := i.m[val]; ok {
		return string(r)
	}
	return val
}

func Nerd() *nerd {
	return &nerd{
		m: map[string]rune{
			// dev
			"Rust":   rune(0xe7a8),
			"Python": rune(0xe73c),

			// seti
			"JavaScript": rune(0xe60c),
			"TypeScript": rune(0xe628),
			"Go":         rune(0xe65e),
			"C":          rune(0xe649),
			"C++":        rune(0xf0646),
			"Reason":     rune(0xe687),
		},
	}
}
