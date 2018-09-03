package matcher

import "testing"

func TestMaskPwd(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"Post https://cargo.caicloudprivatetest.com/login?principal=admin&password=123456: dial tcp: lookup cargo.caicloudprivatetest.com: no such host", "Post https://cargo.caicloudprivatetest.com/login?principal=admin&password=******: dial tcp: lookup cargo.caicloudprivatetest.com: no such host"},
		{"User: admin, password: 123456", "User: admin, password: ******"},
		{"USER:admin, PWD:123456", "USER:admin, PWD:******"},
		{"Compassword=v1.0.0", "Compassword=v1.0.0"},
		{"Pwd: 123456", "Pwd: ******"},
		{"password=123456, and PWD=123456", "password=****** and PWD=******"},
		{"Email:123456@example.com, Pwd: 123456", "Email:123456@example.com, Pwd: ******"},
	}

	for _, c := range cases {
		output := MaskPwd(c.input)
		if output != c.expected {
			t.Errorf(`Expected "%s", but got ""%s"`, c.expected, output)
		}
	}

}

func TestIsIP(t *testing.T) {
	value := "1.1.1.1"
	if !IsIP(value) {
		t.Errorf("%s is IP, expected yes, but got false", value)
	}
	value = "cargo.caicloudprivatetest.com"
	if IsIP(value) {
		t.Errorf("%s is IP, expected no, but got yes", value)
	}
}
