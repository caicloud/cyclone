package slugify

import "testing"

func benchmarkSlugify(b *testing.B, s string) {
	b.StopTimer()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		Slugify(s)
	}
}

func BenchmarkSlugify(b *testing.B) {
	s := `
	天地有正气，杂然赋流形。下则为河岳，上则为日星。于人曰浩然，沛乎塞苍冥。
	But there is something that I must say to my people, who stand on the warm threshold which leads into the palace of justice: In the process of gaining our rightful place, we must not be guilty of wrongful deeds.
	江戸時代末期、「天人（あまんと）」と呼ばれる宇宙人達が襲来した。まもなく地球人と天人との間に十数年にも及ぶ攘夷戦争が勃発。
	12345--`
	benchmarkSlugify(b, s)
}
