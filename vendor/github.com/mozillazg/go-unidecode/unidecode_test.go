package unidecode

import (
	"testing"
	"unicode"
)

type testCase struct {
	input, expect string
}

func testUnidecode(t *testing.T, input, expect string) {
	ret := Unidecode(input)
	check(t, ret, expect)
}

func check(t *testing.T, ret, expect string) {
	if ret != expect {
		t.Errorf("Expected %s, got %s", expect, ret)
	}
}

func TestVersion(t *testing.T) {
	ret := Version()
	expect := "0.1.0"
	check(t, ret, expect)
}

func TestUnidecodeASCII(t *testing.T) {
	for n := 0; n < unicode.MaxASCII; n++ {
		expect := string(rune(n))
		testUnidecode(t, string(rune(n)), expect)
	}
}

func TestUnidecode(t *testing.T) {
	cases := []testCase{
		testCase{"", ""},
		testCase{"abc", "abc"},
		testCase{"北京", "Bei Jing "},
		testCase{"abc北京", "abcBei Jing "},
		testCase{"ネオアームストロングサイクロンジェットアームストロング砲", "neoamusutorongusaikuronzietsutoamusutoronguPao "},
		testCase{"30 𝗄𝗆/𝗁", "30 km/h"},
		testCase{"kožušček", "kozuscek"},
		testCase{"ⓐⒶ⑳⒇⒛⓴⓾⓿", "aA20(20)20.20100"},
		testCase{"Hello, World!", "Hello, World!"},
		testCase{`\n`, `\n`},
		testCase{`北京abc\n`, `Bei Jing abc\n`},
		testCase{`'"\r\n`, `'"\r\n`},
		testCase{"ČŽŠčžš", "CZSczs"},
		testCase{"ア", "a"},
		testCase{"α", "a"},
		testCase{"a", "a"},
		testCase{"ch\u00e2teau", "chateau"},
		testCase{"vi\u00f1edos", "vinedos"},
		testCase{"Efﬁcient", "Efficient"},
		testCase{"příliš žluťoučký kůň pěl ďábelské ódy", "prilis zlutoucky kun pel dabelske ody"},
		testCase{"PŘÍLIŠ ŽLUŤOUČKÝ KŮŇ PĚL ĎÁBELSKÉ ÓDY", "PRILIS ZLUTOUCKY KUN PEL DABELSKE ODY"},
		testCase{"\ua500", ""},
		testCase{"\u1eff", ""},
		testCase{string(0xfffff), ""},
		testCase{"\U0001d5a0", "A"},
		testCase{"\U0001d5c4\U0001d5c6/\U0001d5c1", "km/h"},
		testCase{"\u2124\U0001d552\U0001d55c\U0001d552\U0001d55b \U0001d526\U0001d52a\U0001d51e \U0001d4e4\U0001d4f7\U0001d4f2\U0001d4ec\U0001d4f8\U0001d4ed\U0001d4ee \U0001d4c8\U0001d4c5\u212f\U0001d4b8\U0001d4be\U0001d4bb\U0001d4be\U0001d4c0\U0001d4b6\U0001d4b8\U0001d4be\U0001d4bf\u212f \U0001d59f\U0001d586 \U0001d631\U0001d62a\U0001d634\U0001d622\U0001d637\U0001d626?!", "Zakaj ima Unicode specifikacije za pisave?!"},
	}
	for _, c := range cases {
		testUnidecode(t, c.input, c.expect)
	}
}

func TestUnidecodeConverterA(t *testing.T) {
	v := "THE QUICK BROWN FOX JUMPS OVER THE LAZY DOG 1234567890"
	cases := []string{
		"\uff34\uff28\uff25 \uff31\uff35\uff29\uff23\uff2b \uff22\uff32\uff2f\uff37\uff2e \uff26\uff2f\uff38 \uff2a\uff35\uff2d\uff30\uff33 \uff2f\uff36\uff25\uff32 \uff34\uff28\uff25 \uff2c\uff21\uff3a\uff39 \uff24\uff2f\uff27 \uff11\uff12\uff13\uff14\uff15\uff16\uff17\uff18\uff19\uff10",
		"\U0001d54b\u210d\U0001d53c \u211a\U0001d54c\U0001d540\u2102\U0001d542 \U0001d539\u211d\U0001d546\U0001d54e\u2115 \U0001d53d\U0001d546\U0001d54f \U0001d541\U0001d54c\U0001d544\u2119\U0001d54a \U0001d546\U0001d54d\U0001d53c\u211d \U0001d54b\u210d\U0001d53c \U0001d543\U0001d538\u2124\U0001d550 \U0001d53b\U0001d546\U0001d53e \U0001d7d9\U0001d7da\U0001d7db\U0001d7dc\U0001d7dd\U0001d7de\U0001d7df\U0001d7e0\U0001d7e1\U0001d7d8",
		"\U0001d413\U0001d407\U0001d404 \U0001d410\U0001d414\U0001d408\U0001d402\U0001d40a \U0001d401\U0001d411\U0001d40e\U0001d416\U0001d40d \U0001d405\U0001d40e\U0001d417 \U0001d409\U0001d414\U0001d40c\U0001d40f\U0001d412 \U0001d40e\U0001d415\U0001d404\U0001d411 \U0001d413\U0001d407\U0001d404 \U0001d40b\U0001d400\U0001d419\U0001d418 \U0001d403\U0001d40e\U0001d406 \U0001d7cf\U0001d7d0\U0001d7d1\U0001d7d2\U0001d7d3\U0001d7d4\U0001d7d5\U0001d7d6\U0001d7d7\U0001d7ce",
		"\U0001d47b\U0001d46f\U0001d46c \U0001d478\U0001d47c\U0001d470\U0001d46a\U0001d472 \U0001d469\U0001d479\U0001d476\U0001d47e\U0001d475 \U0001d46d\U0001d476\U0001d47f \U0001d471\U0001d47c\U0001d474\U0001d477\U0001d47a \U0001d476\U0001d47d\U0001d46c\U0001d479 \U0001d47b\U0001d46f\U0001d46c \U0001d473\U0001d468\U0001d481\U0001d480 \U0001d46b\U0001d476\U0001d46e 1234567890",
		"\U0001d4e3\U0001d4d7\U0001d4d4 \U0001d4e0\U0001d4e4\U0001d4d8\U0001d4d2\U0001d4da \U0001d4d1\U0001d4e1\U0001d4de\U0001d4e6\U0001d4dd \U0001d4d5\U0001d4de\U0001d4e7 \U0001d4d9\U0001d4e4\U0001d4dc\U0001d4df\U0001d4e2 \U0001d4de\U0001d4e5\U0001d4d4\U0001d4e1 \U0001d4e3\U0001d4d7\U0001d4d4 \U0001d4db\U0001d4d0\U0001d4e9\U0001d4e8 \U0001d4d3\U0001d4de\U0001d4d6 1234567890",
		"\U0001d57f\U0001d573\U0001d570 \U0001d57c\U0001d580\U0001d574\U0001d56e\U0001d576 \U0001d56d\U0001d57d\U0001d57a\U0001d582\U0001d579 \U0001d571\U0001d57a\U0001d583 \U0001d575\U0001d580\U0001d578\U0001d57b\U0001d57e \U0001d57a\U0001d581\U0001d570\U0001d57d \U0001d57f\U0001d573\U0001d570 \U0001d577\U0001d56c\U0001d585\U0001d584 \U0001d56f\U0001d57a\U0001d572 1234567890",
	}
	for _, c := range cases {
		testUnidecode(t, c, v)
	}
}

func TestUnidecodeConverterB(t *testing.T) {
	v := "the quick brown fox jumps over the lazy dog 1234567890"
	cases := []string{
		"\uff54\uff48\uff45 \uff51\uff55\uff49\uff43\uff4b \uff42\uff52\uff4f\uff57\uff4e \uff46\uff4f\uff58 \uff4a\uff55\uff4d\uff50\uff53 \uff4f\uff56\uff45\uff52 \uff54\uff48\uff45 \uff4c\uff41\uff5a\uff59 \uff44\uff4f\uff47 \uff11\uff12\uff13\uff14\uff15\uff16\uff17\uff18\uff19\uff10",
		"\U0001d565\U0001d559\U0001d556 \U0001d562\U0001d566\U0001d55a\U0001d554\U0001d55c \U0001d553\U0001d563\U0001d560\U0001d568\U0001d55f \U0001d557\U0001d560\U0001d569 \U0001d55b\U0001d566\U0001d55e\U0001d561\U0001d564 \U0001d560\U0001d567\U0001d556\U0001d563 \U0001d565\U0001d559\U0001d556 \U0001d55d\U0001d552\U0001d56b\U0001d56a \U0001d555\U0001d560\U0001d558 \U0001d7d9\U0001d7da\U0001d7db\U0001d7dc\U0001d7dd\U0001d7de\U0001d7df\U0001d7e0\U0001d7e1\U0001d7d8",
		"\U0001d42d\U0001d421\U0001d41e \U0001d42a\U0001d42e\U0001d422\U0001d41c\U0001d424 \U0001d41b\U0001d42b\U0001d428\U0001d430\U0001d427 \U0001d41f\U0001d428\U0001d431 \U0001d423\U0001d42e\U0001d426\U0001d429\U0001d42c \U0001d428\U0001d42f\U0001d41e\U0001d42b \U0001d42d\U0001d421\U0001d41e \U0001d425\U0001d41a\U0001d433\U0001d432 \U0001d41d\U0001d428\U0001d420 \U0001d7cf\U0001d7d0\U0001d7d1\U0001d7d2\U0001d7d3\U0001d7d4\U0001d7d5\U0001d7d6\U0001d7d7\U0001d7ce",
		"\U0001d495\U0001d489\U0001d486 \U0001d492\U0001d496\U0001d48a\U0001d484\U0001d48c \U0001d483\U0001d493\U0001d490\U0001d498\U0001d48f \U0001d487\U0001d490\U0001d499 \U0001d48b\U0001d496\U0001d48e\U0001d491\U0001d494 \U0001d490\U0001d497\U0001d486\U0001d493 \U0001d495\U0001d489\U0001d486 \U0001d48d\U0001d482\U0001d49b\U0001d49a \U0001d485\U0001d490\U0001d488 1234567890",
		"\U0001d4fd\U0001d4f1\U0001d4ee \U0001d4fa\U0001d4fe\U0001d4f2\U0001d4ec\U0001d4f4 \U0001d4eb\U0001d4fb\U0001d4f8\U0001d500\U0001d4f7 \U0001d4ef\U0001d4f8\U0001d501 \U0001d4f3\U0001d4fe\U0001d4f6\U0001d4f9\U0001d4fc \U0001d4f8\U0001d4ff\U0001d4ee\U0001d4fb \U0001d4fd\U0001d4f1\U0001d4ee \U0001d4f5\U0001d4ea\U0001d503\U0001d502 \U0001d4ed\U0001d4f8\U0001d4f0 1234567890",
		"\U0001d599\U0001d58d\U0001d58a \U0001d596\U0001d59a\U0001d58e\U0001d588\U0001d590 \U0001d587\U0001d597\U0001d594\U0001d59c\U0001d593 \U0001d58b\U0001d594\U0001d59d \U0001d58f\U0001d59a\U0001d592\U0001d595\U0001d598 \U0001d594\U0001d59b\U0001d58a\U0001d597 \U0001d599\U0001d58d\U0001d58a \U0001d591\U0001d586\U0001d59f\U0001d59e \U0001d589\U0001d594\U0001d58c 1234567890",
	}
	for _, c := range cases {
		testUnidecode(t, c, v)
	}
}
