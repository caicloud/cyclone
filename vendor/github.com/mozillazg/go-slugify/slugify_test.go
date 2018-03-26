package slugify

import "testing"
import "log"

type testCase struct {
	input, expect string
}

func testSlugify(t *testing.T, input, expect string) {
	ret := Slugify(input)
	check(t, ret, expect)
}

func check(t *testing.T, ret, expect string) {
	if ret != expect {
		t.Errorf("Expected %v, got %v", expect, ret)
	}
}

func TestVersion(t *testing.T) {
	ret := Version()
	expect := "0.2.0"
	check(t, ret, expect)
}

func TestSlugify(t *testing.T) {
	cases := []testCase{
		{"", ""},
		{"abc", "abc"},
		{"abc234", "abc234"},
		{"This is a test ---", "this-is-a-test"},
		{"___This is a test___", "this-is-a-test"},
		{"This -- is a ## test ---", "this-is-a-test"},
		{"北京kožušček", "bei-jing-kozuscek"},
		{"Nín hǎo. Wǒ shì zhōng guó rén", "nin-hao-wo-shi-zhong-guo-ren"},
		{`C\'est déjà l\'été.`, "c-est-deja-l-ete"},
	}
	for _, c := range cases {
		log.Printf("input %v, expect %v", c.input, c.expect)
		testSlugify(t, c.input, c.expect)
	}
}
