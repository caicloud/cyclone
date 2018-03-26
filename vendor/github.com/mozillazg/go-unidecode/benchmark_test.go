package unidecode

import "testing"

func benchmarkUnidecode(b *testing.B, s string) {
	b.StopTimer()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		Unidecode(s)
	}
}

func BenchmarkUnidecode1500(b *testing.B) {
	chinese := []rune(`
	余囚北庭，坐一土室。室广八尺，深可四寻。单扉低小，白间短窄，污下而幽暗。当此夏日，诸气萃然：雨潦四集，浮动床几，时则为水气；涂泥半朝，蒸沤历澜，时则为土气；乍晴暴热，风道四塞，时则为日气；檐阴薪爨，助长炎虐，时则为火气；仓腐寄顿，陈陈逼人，时则为米气；骈肩杂遝，腥臊汗垢，时则为人气；或圊溷、或毁尸、或腐鼠，恶气杂出，时则为秽气。叠是数气，当之者鲜不为厉。而予以孱弱，俯仰其间，于兹二年矣，幸而无恙，是殆有养致然尔。然亦安知所养何哉？孟子曰：‘吾善养吾浩然之气。’彼气有七，吾气有一，以一敌七，吾何患焉！况浩然者，乃天地之正气也，作正气歌一首。
	天地有正气，杂然赋流形。下则为河岳，上则为日星。于人曰浩然，沛乎塞苍冥。
	皇路当清夷，含和吐明庭。时穷节乃见，一一垂丹青。在齐太史简，在晋董狐笔。
	在秦张良椎，在汉苏武节。为严将军头，为嵇侍中血。为张睢阳齿，为颜常山舌。
	或为辽东帽，清操厉冰雪。或为出师表，鬼神泣壮烈。或为渡江楫，慷慨吞胡羯。
	或为击贼笏，逆竖头破裂。是气所磅礡，凛烈万古存。当其贯日月，生死安足论。
	地维赖以立，天柱赖以尊。三纲实系命，道义为之根。嗟予遘阳九，隶也实不力。
	楚囚缨其冠，传车送穷北。鼎镬甘如饴，求之不可得。阴房阒鬼火，春院閟天黑。
	牛骥同一皂，鸡栖凤凰食。一朝蒙雾露，分作沟中瘠。如此再寒暑，百沴自辟易。
	嗟哉沮洳场，为我安乐国。岂有他缪巧，阴阳不能贼。顾此耿耿在，仰视浮云白。
	悠悠我心悲，苍天曷有极。哲人日已远，典型在夙昔。风檐展书读，古道照颜色。
	`)[:500]
	english := []rune(`
	But there is something that I must say to my people, who stand on the warm threshold which leads into the palace of justice: In the process of gaining our rightful place, we must not be guilty of wrongful deeds. Let us not seek to satisfy our thirst for freedom by drinking from the cup of bitterness and hatred. We must forever conduct our struggle on the high plane of dignity and discipline. We must not allow our creative protest to degenerate into physical violence. Again and again, we must rise to the majestic heights of meeting physical force with soul force.
	`)[:500]
	japanese := []rune(`
	江戸時代末期、「天人（あまんと）」と呼ばれる宇宙人達が襲来した。まもなく地球人と天人との間に十数年にも及ぶ攘夷戦争が勃発。数多くの侍・攘夷志士が天人との戦争に参加した。しかし天人の絶大な力を見て弱腰になっていた幕府は、天人の侵略をあっさりと受け入れ開国してしまう。そして幕府は天人による傀儡政権となり、天人達が我が物顔で江戸の街を闊歩するようになった。一方、国・主君のために天人と戦った攘夷志士達は弾圧の対象となり、他の侍達もその多くが廃刀令によって刀を失い、力を奪われていった。
	天人の襲来から20年後、剣術道場の跡取りの志村新八は剣術を生かす道も無く、意に沿わないアルバイトで姉である志村妙となんとか生計を立てていた。そんな新八の前に風変わりな一人の侍が現れる。未だに変わらない侍魂を持った男、その名も坂田銀時。銀時の男気に惹かれた新八は、侍の魂を学ぶために彼の営業する万事屋で働き出す。やがて万事屋には、戦闘種族である夜兎族の神楽・巨大犬の定春などが転がり込んでくる。
	そして万事屋ゆえに江戸のあらゆる依頼事に首を突っ込むようになった銀時達は、江戸の治安を預かる真選組・かつて銀時の盟友であった侍達等、様々な人間達と関わり合っていく事になる。
`)[:500]
	r := []rune{}
	r = append(r, chinese...)
	r = append(r, english...)
	r = append(r, japanese...)
	s := string(r)
	benchmarkUnidecode(b, s)
}
