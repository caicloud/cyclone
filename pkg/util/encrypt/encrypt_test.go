/*
Copyright 2017 caicloud authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package encrypt

import (
	"fmt"
	"testing"
)

func TestEncrypt(t *testing.T) {
	testCases := map[string]struct {
		src string
		key string
	}{
		"key length 16": {
			"123456",
			"1234567812345678",
		},
		"key length 24": {
			"password",
			"123456781234567812345678",
		},
		"key length 32": {
			"password",
			"12345678123456781234567812345678",
		},
		"long content": {
			"passwordPASSWORDpasswordPASSWORD",
			"1234567812345678",
		},
		"short content": {
			"pwd",
			"1234567812345678",
		},
		"content with special character": {
			"1Ab~%……&*（）￥#@？、，./{}[]",
			"1234567812345678",
		},
	}

	for d, tc := range testCases {
		dst, err := Encrypt(tc.src, tc.key)
		if err != nil {
			t.Fatalf("%s: fail to encrypt %s: %v", d, tc.src, err)
		}

		fmt.Println(dst)
		src, err := Decrypt(dst, tc.key)
		if err != nil {
			t.Fatalf("%s: fail to decrypt %s: %v", d, tc.src, err)
		}

		if src != tc.src {
			t.Fatalf("%s: expected %s, but %s", d, tc.src, src)
		}
	}

	errTestCases := map[string]struct {
		src string
		key string
	}{
		"too short key": {
			"password",
			"1234567",
		},
		"too long key": {
			"password",
			"1234567812345678123456781234567812345678",
		},
		"wrong key length": {
			"password",
			"123456781234567",
		},
	}

	for d, tc := range errTestCases {
		_, err := Encrypt(tc.src, tc.key)
		if err == nil {
			t.Fatalf("%s: fail to encrypt %s: %v", d, tc.src, err)
		}
	}
}

func TestDecrypt(t *testing.T) {
	testCases := map[string]struct {
		src string
		key string
	}{
		"key length 16": {
			"W3OkyHNTst7eiCsy3Bu+So5hq3rx6DrO4ovsgr/ACdQJLZ74rkdv2pDK6t7rcRJWklpWeTOl8Q==",
			"1234567812345678",
		},
		"key length 24": {
			"Ciz5+p0JpbzAiGvNbXsLnNCSH9KpTeCB+HpXdUHG72Lhu8x2IQKDJYVI0UmUYbMQ",
			"123456781234567812345678",
		},
		"key length 32": {
			"UWqdq9y+XaNfAq0zh2V+jIXmFA==",
			"12345678123456781234567812345678",
		},
		"long content": {
			"+QnJxqRC/ENNZxBkdW/lQiUTM2KzUKGk7223GoPe++hMDcrnhmB6vUlCYn5wvnbjlB9V80i90w==",
			"1234567812345678",
		},
		"short content": {
			"kW6tzltgDSNhhoeT1lb3VfVvMjJyuQih",
			"1234567812345678",
		},
		"content with special character": {
			"BwzFxyHKzMejUIAl/JvRe4PBPRY3SKmG",
			"1234567812345678",
		},
	}

	for d, tc := range testCases {
		_, err := Decrypt(tc.src, tc.key)
		if err != nil {
			t.Fatalf("%s: fail to decrypt %s: %v", d, tc.src, err)
		}
	}

	errTestCases := map[string]struct {
		src string
		key string
	}{
		"too short key": {
			"password",
			"1234567",
		},
		"too long key": {
			"password",
			"1234567812345678123456781234567812345678",
		},
		"wrong key length": {
			"password",
			"123456781234567",
		},
		"too short content": {
			"password",
			"1234567812345678",
		},
	}

	for d, tc := range errTestCases {
		_, err := Decrypt(tc.src, tc.key)
		if err == nil {
			t.Fatalf("%s: fail to encrypt %s: %v", d, tc.src, err)
		}
	}
}
