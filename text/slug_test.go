package text_test

import (
	"testing"

	"github.com/lucasepe/toolbox/text"
)

var tests = []struct{ in, out string }{
	{"simple test", "simple-test"},
	{"I'm go developer", "i-m-go-developer"},
	{"Simples código em go", "simples-codigo-em-go"},
	{"日本語の手紙をテスト", "日本語の手紙をテスト"},
	{"--->simple test<---", "simple-test"},
}

func TestSlugify(t *testing.T) {
	for _, test := range tests {
		if out := text.Slugify(test.in); out != test.out {
			t.Errorf("%q: %q != %q", test.in, out, test.out)
		}
	}
}

func TestSlugifyf(t *testing.T) {
	for _, test := range tests {
		t.Run(test.out, func(t *testing.T) {
			if out := text.Slugifyf("%s", test.in); out != test.out {
				t.Errorf("%q: %q != %q", test.in, out, test.out)
			}
		})
	}
}
