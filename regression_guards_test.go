package recommender

import (
	"strings"
	"testing"
)

// Regression guards for bugs found in production audits.
// Each test names the failure mode it prevents — if scoring is "simplified" later,
// these should fail loudly.

func TestRegressionSeriesBaseStripsLiveActionBranding(t *testing.T) {
	// Without this, commercial base becomes "one piece live action" and loses
	// franchise peers to Smallville/Knight Rider on pure genre overlap.
	got := seriesBase("ONE PIECE (Live Action) (Season 2)")
	want := "one piece"
	if got != want {
		t.Fatalf("seriesBase(live-action OP) = %q; want %q", got, want)
	}
	if commercialSeriesBase("ONE PIECE (Live Action) (Season 2)") != commercialSeriesBase("One Piece: Curse Of The Sacred Sword") {
		t.Fatalf("commercial bases must match across LA vs anime: %q vs %q",
			commercialSeriesBase("ONE PIECE (Live Action) (Season 2)"),
			commercialSeriesBase("One Piece: Curse Of The Sacred Sword"))
	}
	if commercialSeriesBase("Spider-Man: Homecoming") != commercialSeriesBase("Spider-Man 2") {
		t.Fatalf("spider commercial bases should match: %q vs %q",
			commercialSeriesBase("Spider-Man: Homecoming"),
			commercialSeriesBase("Spider-Man 2"))
	}
}

func TestRegressionTalesSubtitleNotFranchiseIdentity(t *testing.T) {
	// Stranger Things: Tales From '85 was matching Trollhunters: Tales of Arcadia
	// via weak single-token subtitle root "tales".
	if !isWeakContainmentToken("tales") || !isWeakContainmentToken("story") {
		t.Fatal("media-generic subtitle words must be weak containment tokens")
	}
	if !isRootStopword("tales") {
		t.Fatal("tales must be a root stopword so subtitles do not become franchise roots")
	}

	st := Movie{
		Name: "Cậu Bé Mất Tích: Chuyện Năm 85", OriginName: "Stranger Things: Tales From '85",
		Slug: "cau-be-mat-tich-chuyen-nam-85",
		Genres: []string{"khoa-hoc", "vien-tuong", "hoathinh"}, Country: "Âu Mỹ",
	}
	troll := Movie{
		Name: "Thợ Săn Yêu Tinh: Truyền Thuyết Arcadia (Phần 1)",
		OriginName: "Trollhunters: Tales of Arcadia (Season 1)",
		Slug: "tho-san-yeu-tinh-truyen-thuyet-arcadia-phan-1",
		Genres: []string{"hanh-dong", "phieu-luu", "hoat-hinh"}, Country: "Âu Mỹ",
	}
	if sc := ScoreSeriesMatch(st, troll); sc != 0 {
		t.Fatalf("Tales-subtitle collision: ScoreSeriesMatch=%f want 0 (ST vs Trollhunters)", sc)
	}
	res := Recommend(st, []Movie{troll}, UserContext{})
	if len(res.SameSeries) != 0 {
		t.Fatalf("same_series must not include Trollhunters, got %v", slugs(res.SameSeries))
	}
}

func TestRegressionSensitiveNotMatureAnimationBlurb(t *testing.T) {
	// Invincible descriptions say "dành cho người lớn" — must NOT flag as adult.
	inv := Movie{
		Name: "Bất Khả Chiến Bại (Phần 1)", OriginName: "Invincible (Season 1)",
		Slug: "bat-kha-chien-bai-phan-1",
		Genres: []string{"chinh-kich", "phieu-luu", "vien-tuong"},
		Content: "Invincible là series hoạt hình siêu anh hùng dành cho người lớn được chuyển thể từ truyện tranh.",
	}
	if IsSensitiveContent(inv) {
		t.Fatal("mature-animation blurb must not mark Invincible as sensitive")
	}
	// Real adult title markers must still trip.
	adult := Movie{
		Name: "Nhật Ký Cô Nàng Nghiện Sex", Slug: "nhat-ky-co-nang-nghien-sex",
		Genres: []string{"tinh-cam"},
	}
	if !IsSensitiveContent(adult) {
		t.Fatal("title containing sex must remain sensitive")
	}
	adult2 := Movie{
		Name: "Người Đàn Bà Cuồng Dâm: Phần 1", Slug: "nguoi-dan-ba-cuong-dam-phan-1",
		Genres: []string{"chinh-kich"},
	}
	if !IsSensitiveContent(adult2) {
		t.Fatal("cuồng dâm title must remain sensitive")
	}
}

func TestRegressionDemonSlayerNotMidTitleContainment(t *testing.T) {
	// "demon slayer" mid-string in "immortal demon slayer (wukong)" must not same_series.
	ds := Movie{
		Name: "Thanh Gươm Diệt Quỷ (Phần 1)", OriginName: "Demon Slayer (Season 1)",
		Slug: "thanh-guom-diet-quy-phan-1", Genres: []string{"hanh-dong", "phieu-luu"}, Country: "nhat-ban",
	}
	wukong := Movie{
		Name: "Ngộ Không Kỳ Truyện", OriginName: "Immortal Demon Slayer (Wukong)",
		Slug: "ngo-khong-ky-truyen", Genres: []string{"hanh-dong", "vien-tuong"}, Country: "trung-quoc",
	}
	if sc := ScoreSeriesMatch(ds, wukong); sc != 0 {
		t.Fatalf("Wukong/Immortal Demon Slayer must not same_series with Demon Slayer, score=%f", sc)
	}
	// True sequel still matches.
	ds2 := Movie{
		Name: "Thanh Gươm Diệt Quỷ (Phần 2)", OriginName: "Demon Slayer (Season 2)",
		Slug: "thanh-guom-diet-quy-phan-2", Genres: []string{"hanh-dong", "phieu-luu"}, Country: "nhat-ban",
	}
	if ScoreSeriesMatch(ds, ds2) < 10 {
		t.Fatalf("true Demon Slayer seasons must still match, score=%f", ScoreSeriesMatch(ds, ds2))
	}
}

func TestRegressionInvincibleNotMedicFattyOrBareChineseTitle(t *testing.T) {
	inv4 := Movie{
		Name: "Bất Khả Chiến Bại (Phần 4)", OriginName: "Invincible (Season 4)",
		Slug: "bat-kha-chien-bai-phan-4",
		Genres: []string{"chinh-kich", "phieu-luu", "vien-tuong", "tam-ly"}, Country: "canada",
	}
	bads := []Movie{
		{Name: "Tình Nghĩa Giang Hồ", OriginName: "Invincible Medic",
			Slug: "tinh-nghia-giang-ho", Genres: []string{"chinh-kich"}, Country: "trung-quoc"},
		{Name: "Béo bất khả chiến bại", OriginName: "Invincible Fatty",
			Slug: "beo-bat-kha-chien-bai", Genres: []string{"hai-huoc"}, Country: "trung-quoc"},
		{Name: "Trần Tường 6h30", OriginName: "Invincible",
			Slug: "tran-tuong-6h30-nguoi-me-quyen-anh", Genres: []string{"hanh-dong"}, Country: "trung-quoc"},
	}
	for _, bad := range bads {
		if sc := ScoreSeriesMatch(inv4, bad); sc != 0 {
			t.Fatalf("Invincible impostor %s same_series score=%f want 0", bad.Slug, sc)
		}
	}
	good := Movie{
		Name: "Bất Khả Chiến Bại (Phần 1)", OriginName: "Invincible (Season 1)",
		Slug: "bat-kha-chien-bai-phan-1",
		Genres: inv4.Genres, Country: "canada",
	}
	if ScoreSeriesMatch(inv4, good) < 10 {
		t.Fatalf("true Invincible season must match, score=%f", ScoreSeriesMatch(inv4, good))
	}
}

func TestRegressionJujutsuMissingHoatHinhStillAnime(t *testing.T) {
	// Catalog often omits hoat-hinh on JJK seasons — must still separate from US live series.
	jjk := Movie{
		Name: "Chú Thuật Hồi Chiến (Phần 2)", OriginName: "Jujutsu Kaisen (Season 2)",
		Slug: "chu-thuat-hoi-chien-phan-2",
		Genres: []string{"hanh-dong", "phieu-luu", "khoa-hoc", "vien-tuong", "tam-ly"}, // no hoat-hinh
		Country: "nhat-ban",
	}
	if animationFormat(jjk) != "anime" {
		t.Fatalf("JJK without hoat-hinh must classify as anime, got %q", animationFormat(jjk))
	}
	small := Movie{
		Name: "Thị Trấn Smallville (Phần 4)", OriginName: "Smallville (Season 4)",
		Slug: "thi-tran-smallville-phan-4",
		Genres: []string{"khoa-hoc", "vien-tuong", "hanh-dong", "phieu-luu", "chinh-kich", "tam-ly", "series"},
		Country: "Âu Mỹ",
	}
	if sc := ScoreSimilarContent(jjk, small); sc > 0 {
		t.Fatalf("JJK must not positively rank Smallville, similar=%f", sc)
	}
}

func TestRegressionOnePieceLAFranchiseBeatsUSLiveNoise(t *testing.T) {
	opLA := Movie{
		Name: "Đảo Hải Tặc (Live Action) (Phần 2)", OriginName: "ONE PIECE (Live Action) (Season 2)",
		Slug: "dao-hai-tac-live-action-phan-2",
		Genres: []string{"hanh-dong", "phieu-luu", "khoa-hoc", "vien-tuong", "series"},
		Country: "Âu Mỹ",
	}
	opAnime := Movie{
		Name: "Đảo Hải Tặc 5: Lời Nguyền Thành Kiếm", OriginName: "One Piece: Curse Of The Sacred Sword",
		Slug: "dao-hai-tac-5-loi-nguyen-thanh-kiem",
		Genres: []string{"hoat-hinh", "hanh-dong", "phieu-luu"}, Country: "Nhật Bản",
	}
	noise := []Movie{
		{Slug: "thi-tran-smallville-phan-4", Name: "Smallville S4", OriginName: "Smallville (Season 4)",
			Genres: []string{"khoa-hoc", "vien-tuong", "hanh-dong", "phieu-luu", "chinh-kich", "series"}, Country: "Âu Mỹ"},
		{Slug: "doi-dac-nhiem-shield-phan-2", Name: "SHIELD S2", OriginName: "Agents of S.H.I.E.L.D. (Season 2)",
			Genres: []string{"hanh-dong", "phieu-luu", "khoa-hoc", "vien-tuong", "series"}, Country: "Âu Mỹ"},
		{Slug: "hiep-si-xe-den-phan-1", Name: "Knight Rider", OriginName: "Knight Rider (Season 1)",
			Genres: []string{"hanh-dong", "phieu-luu", "khoa-hoc", "series"}, Country: "Âu Mỹ"},
	}
	opScore := ScoreSimilarContent(opLA, opAnime)
	if opScore <= 0 {
		t.Fatalf("OP anime peer must score positive for LA target, got %f", opScore)
	}
	for _, n := range noise {
		ns := ScoreSimilarContent(opLA, n)
		if opScore <= ns {
			t.Fatalf("OP franchise %f must beat %s %f", opScore, n.Slug, ns)
		}
	}
	// Live-action must not share same_series with anime movie.
	if ScoreSeriesMatch(opLA, opAnime) != 0 {
		t.Fatalf("OP LA vs anime movie same_series must be 0, got %f", ScoreSeriesMatch(opLA, opAnime))
	}
	res := Recommend(opLA, append(noise, opAnime), UserContext{})
	if len(res.SimilarContent) == 0 || !strings.Contains(res.SimilarContent[0].Slug, "dao-hai-tac") && !strings.Contains(res.SimilarContent[0].Slug, "one-piece") {
		t.Fatalf("Recommend similar[0] must be OP franchise, got %v", slugs(res.SimilarContent))
	}
}

func TestRegressionSpiderVerseSimilarRanksLiveFranchise(t *testing.T) {
	verse := Movie{
		Name: "Người Nhện: Du Hành Vũ Trụ Nhện", OriginName: "Spider-Man: Across The Spider-Verse",
		Slug: "nguoi-nhen-du-hanh-vu-tru-nhen",
		Genres: []string{"hanh-dong", "phieu-luu", "khoa-hoc", "vien-tuong", "hoathinh"}, Country: "Âu Mỹ",
	}
	live := Movie{
		Name: "Người Nhện 2", OriginName: "Spider-Man 2", Slug: "nguoi-nhen-2",
		Genres: []string{"hanh-dong", "phieu-luu", "khoa-hoc", "vien-tuong"}, Country: "Âu Mỹ",
	}
	cartoonNoise := Movie{
		Name: "Huyền Thoại Korra (Phần 1)", OriginName: "The Legend Of Korra (Season 1)",
		Slug: "huyen-thoai-korra-phan-1",
		Genres: []string{"hanh-dong", "phieu-luu", "khoa-hoc", "vien-tuong", "hoat-hinh"}, Country: "Âu Mỹ",
	}
	if ScoreSimilarContent(verse, live) <= ScoreSimilarContent(verse, cartoonNoise) {
		t.Fatalf("live Spider-Man (%f) must outrank Korra (%f) for Spider-Verse",
			ScoreSimilarContent(verse, live), ScoreSimilarContent(verse, cartoonNoise))
	}
	// Animated continuity stays same_series; live does not.
	into := Movie{
		Name: "Người Nhện: Vũ Trụ Mới", OriginName: "Spider-Man: Into The Spider-Verse",
		Slug: "nguoi-nhen-vu-tru-moi",
		Genres: []string{"hanh-dong", "hoat-hinh"}, Country: "Âu Mỹ",
	}
	if ScoreSeriesMatch(verse, into) < 10 {
		t.Fatalf("Into Spider-Verse must be same_series, score=%f", ScoreSeriesMatch(verse, into))
	}
	if ScoreSeriesMatch(verse, live) != 0 {
		t.Fatalf("live Spider-Man must not be same_series for animated target, score=%f", ScoreSeriesMatch(verse, live))
	}
}

func TestRegressionShortlistNeverFullCatalogScan(t *testing.T) {
	// Thin franchise + large filler pool must not fall back to scoring all fillers.
	target := Movie{
		Name: "Bất Khả Chiến Bại (Phần 4)", OriginName: "Invincible (Season 4)",
		Slug: "bat-kha-chien-bai-phan-4",
		Genres: []string{"chinh-kich", "phieu-luu", "vien-tuong", "tam-ly"}, Country: "canada",
	}
	cands := make([]Movie, 0, 3000)
	for i := 0; i < 2500; i++ {
		cands = append(cands, Movie{
			Slug: "filler-" + itoaCorpus(i), Name: "Filler", OriginName: "Filler Film",
			Genres: []string{"hai-huoc"}, Country: "trung-quoc",
		})
	}
	cands = append(cands, Movie{
		Slug: "bat-kha-chien-bai-phan-1", Name: "Inv S1", OriginName: "Invincible (Season 1)",
		Genres: target.Genres, Country: "canada",
	})
	out := shortlistCandidates(target, cands, UserContext{})
	if len(out) >= len(cands) {
		t.Fatalf("shortlist fell back to full set: %d", len(out))
	}
	if len(out) > 500 {
		t.Fatalf("shortlist too large for thin franchise: %d", len(out))
	}
	// Peer must survive shortlist → Recommend
	res := Recommend(target, cands, UserContext{})
	if !hasSlug(res.SameSeries, "bat-kha-chien-bai-phan-1") {
		t.Fatalf("shortlist dropped true Invincible peer, same=%v", slugs(res.SameSeries))
	}
}

func TestRegressionRootContainsRequiresPrefixNotMidTitle(t *testing.T) {
	// Mid-title "demon slayer" inside "immortal demon slayer" is NOT a prefix match.
	if rootContainsAsPrefix("demon slayer", "immortal demon slayer") {
		t.Fatal("mid-title containment must not count as prefix")
	}
	if !rootContainsAsPrefix("larva", "larva island") {
		t.Fatal("true franchise prefix larva ⊂ larva island must hold")
	}
	if !rootContainsAsPrefix("one piece", "one piece curse of the sacred sword") {
		t.Fatal("one piece prefix of longer title must hold")
	}
}
