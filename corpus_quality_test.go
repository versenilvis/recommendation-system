package recommender

import (
	"strings"
	"testing"
)

// Corpus fixtures are faithful to real catalog rows (see phim-hay data.db audit).
// Tests drive the shipped Recommend entry point with property assertions.

func corpusPool() []Movie {
	return []Movie{
		// Spider-Verse + live
		{Slug: "nguoi-nhen-du-hanh-vu-tru-nhen", Name: "Người Nhện: Du Hành Vũ Trụ Nhện", OriginName: "Spider-Man: Across The Spider-Verse", Genres: []string{"hanh-dong", "phieu-luu", "khoa-hoc", "vien-tuong", "tam-ly", "hoathinh"}, Country: "Âu Mỹ", Year: 2023},
		{Slug: "nguoi-nhen-vu-tru-moi", Name: "Người Nhện: Vũ Trụ Mới", OriginName: "Spider-Man: Into The Spider-Verse", Genres: []string{"hanh-dong", "phieu-luu", "khoa-hoc", "vien-tuong", "tam-ly", "hoat-hinh"}, Country: "Âu Mỹ", Year: 2018},
		{Slug: "nguoi-nhen-2", Name: "Người Nhện 2", OriginName: "Spider-Man 2", Genres: []string{"hanh-dong", "phieu-luu", "khoa-hoc", "vien-tuong", "tam-ly"}, Country: "Âu Mỹ", Year: 2004, Actors: []string{"Tobey Maguire", "Kirsten Dunst"}, Directors: []string{"Sam Raimi"}},
		{Slug: "nguoi-nhen-3", Name: "Người Nhện 3", OriginName: "Spider-Man 3", Genres: []string{"hanh-dong", "phieu-luu", "khoa-hoc", "vien-tuong", "tam-ly"}, Country: "au-my", Year: 2007, Actors: []string{"Tobey Maguire", "Kirsten Dunst"}, Directors: []string{"Sam Raimi"}},
		{Slug: "nguoi-nhen-2002", Name: "Người Nhện", OriginName: "Spider-Man", Genres: []string{"hanh-dong", "khoa-hoc", "vien-tuong"}, Country: "Âu Mỹ", Year: 2002, Actors: []string{"Tobey Maguire"}, Directors: []string{"Sam Raimi"}},
		{Slug: "nguoi-nhen-tro-ve-nha", Name: "Người Nhện: Trở Về Nhà", OriginName: "Spider-Man: Homecoming", Genres: []string{"hanh-dong", "phieu-luu", "vien-tuong"}, Country: "Âu Mỹ", Year: 2017, Actors: []string{"Tom Holland"}, Directors: []string{"Jon Watts"}},
		// JJK (missing hoat-hinh on seasons)
		{Slug: "chu-thuat-hoi-chien-phan-1", Name: "Chú Thuật Hồi Chiến (Phần 1)", OriginName: "Jujutsu Kaisen (Season 1)", Genres: []string{"hanh-dong", "phieu-luu", "khoa-hoc", "vien-tuong", "tam-ly"}, Country: "nhat-ban", Year: 2020},
		{Slug: "chu-thuat-hoi-chien-phan-2", Name: "Chú Thuật Hồi Chiến (Phần 2)", OriginName: "Jujutsu Kaisen (Season 2)", Genres: []string{"hanh-dong", "phieu-luu", "khoa-hoc", "vien-tuong", "tam-ly"}, Country: "nhat-ban", Year: 2023},
		{Slug: "chu-thuat-hoi-chien-phan-3", Name: "Chú Thuật Hồi Chiến (Phần 3)", OriginName: "Jujutsu Kaisen (Season 3)", Genres: []string{"hanh-dong", "phieu-luu", "khoa-hoc", "vien-tuong", "tam-ly", "hoathinh"}, Country: "Nhật Bản", Year: 2025},
		// Live-action noise that must NOT appear for anime
		{Slug: "thi-tran-smallville-phan-4", Name: "Thị Trấn Smallville (Phần 4)", OriginName: "Smallville (Season 4)", Genres: []string{"khoa-hoc", "vien-tuong", "hanh-dong", "phieu-luu", "chinh-kich", "tam-ly", "series"}, Country: "Âu Mỹ"},
		// Invincible family + impostors
		{Slug: "bat-kha-chien-bai-phan-1", Name: "Bất Khả Chiến Bại (Phần 1)", OriginName: "Invincible (Season 1)", Genres: []string{"chinh-kich", "phieu-luu", "vien-tuong", "tam-ly"}, Country: "canada", Content: "Invincible là series hoạt hình siêu anh hùng dành cho người lớn."},
		{Slug: "bat-kha-chien-bai-phan-2", Name: "Bất Khả Chiến Bại (Phần 2)", OriginName: "Invincible (Season 2)", Genres: []string{"chinh-kich", "phieu-luu", "vien-tuong", "tam-ly"}, Country: "canada"},
		{Slug: "bat-kha-chien-bai-phan-3", Name: "Bất Khả Chiến Bại (Phần 3)", OriginName: "Invincible (Season 3)", Genres: []string{"chinh-kich", "phieu-luu", "vien-tuong", "tam-ly"}, Country: "canada"},
		{Slug: "bat-kha-chien-bai-phan-4", Name: "Bất Khả Chiến Bại (Phần 4)", OriginName: "Invincible (Season 4)", Genres: []string{"chinh-kich", "phieu-luu", "vien-tuong", "tam-ly"}, Country: "canada"},
		{Slug: "invincible-nguon-goc-atom-eve", Name: "Invincible: Nguồn gốc Atom Eve", OriginName: "Invincible: Presenting Atom Eve", Genres: []string{"chinh-kich", "hanh-dong", "hoathinh"}, Country: "Âu Mỹ"},
		{Slug: "tinh-nghia-giang-ho", Name: "Tình Nghĩa Giang Hồ", OriginName: "Invincible Medic", Genres: []string{"chinh-kich"}, Country: "trung-quoc"},
		{Slug: "beo-bat-kha-chien-bai", Name: "Béo bất khả chiến bại", OriginName: "Invincible Fatty", Genres: []string{"hai-huoc"}, Country: "trung-quoc"},
		{Slug: "tran-tuong-6h30-nguoi-me-quyen-anh", Name: "Trần Tường 6h30: Người Mẹ Quyền Anh", OriginName: "Invincible", Genres: []string{"hanh-dong", "gia-dinh"}, Country: "trung-quoc"},
		// Demon Slayer + false friend
		{Slug: "thanh-guom-diet-quy-phan-1-kamado-tanjiro-lap-chi", Name: "Thanh Gươm Diệt Quỷ (Phần 1)", OriginName: "Demon Slayer (Season 1) (Tanjiro Kamado, Unwavering Resolve Arc)", Genres: []string{"hanh-dong", "phieu-luu"}, Country: "nhat-ban"},
		{Slug: "thanh-guom-diet-quy-phan-2-chuyen-tau-vo-tan", Name: "Thanh Gươm Diệt Quỷ (Phần 2)", OriginName: "Demon Slayer (Season 2) (Mugen Train Arc)", Genres: []string{"hanh-dong", "phieu-luu"}, Country: "nhat-ban"},
		{Slug: "thanh-guom-diet-quy-vo-han-thanh", Name: "Thanh Gươm Diệt Quỷ: Vô Hạn Thành", OriginName: "Demon Slayer: Kimetsu no Yaiba Infinity Castle", Genres: []string{"hanh-dong", "vien-tuong", "hoathinh"}, Country: "Nhật Bản"},
		{Slug: "ngo-khong-ky-truyen", Name: "Ngộ Không Kỳ Truyện", OriginName: "Immortal Demon Slayer (Wukong)", Genres: []string{"vien-tuong", "hanh-dong"}, Country: "trung-quoc"},
		// One Piece live vs anime
		{Slug: "dao-hai-tac-live-action-phan-2", Name: "Đảo Hải Tặc (Live Action) (Phần 2)", OriginName: "ONE PIECE (Live Action) (Season 2)", Genres: []string{"hanh-dong", "phieu-luu", "series"}, Country: "Âu Mỹ"},
		{Slug: "dao-hai-tac-5-loi-nguyen-thanh-kiem", Name: "Đảo Hải Tặc 5: Lời Nguyền Thành Kiếm", OriginName: "One Piece: Curse Of The Sacred Sword", Genres: []string{"hoat-hinh", "hanh-dong"}, Country: "Nhật Bản"},
		// Adult decoys for guest rails
		{Slug: "nhat-ky-co-nang-nghien-sex", Name: "Nhật Ký Cô Nàng Nghiện Sex", OriginName: "Sex Diary", Genres: []string{"tinh-cam", "chinh-kich"}, Country: "Tây Ban Nha"},
		{Slug: "nguoi-dan-ba-cuong-dam-phan-1", Name: "Người Đàn Bà Cuồng Dâm: Phần 1", OriginName: "Lady Chatterley", Genres: []string{"chinh-kich"}, Country: "Anh"},
		// Filler for shortlist
		{Slug: "filler-cartoon", Name: "Random Cartoon", OriginName: "Random Cartoon Show", Genres: []string{"hoat-hinh", "hanh-dong"}, Country: "Âu Mỹ"},
		{Slug: "filler-jp", Name: "Random JP Drama", OriginName: "Random Japanese Drama", Genres: []string{"chinh-kich", "tam-ly"}, Country: "nhat-ban"},
	}
}

func without(pool []Movie, slug string) []Movie {
	out := make([]Movie, 0, len(pool)-1)
	for _, m := range pool {
		if m.Slug != slug {
			out = append(out, m)
		}
	}
	return out
}

func find(pool []Movie, slug string) Movie {
	for _, m := range pool {
		if m.Slug == slug {
			return m
		}
	}
	panic("missing " + slug)
}

func slugs(ms []Movie) []string {
	out := make([]string, len(ms))
	for i, m := range ms {
		out[i] = m.Slug
	}
	return out
}

func hasSlug(ms []Movie, slug string) bool {
	for _, m := range ms {
		if m.Slug == slug {
			return true
		}
	}
	return false
}

func hasAnySlugPrefix(ms []Movie, prefixes ...string) bool {
	for _, m := range ms {
		for _, p := range prefixes {
			if strings.HasPrefix(m.Slug, p) || strings.Contains(m.Slug, p) {
				return true
			}
		}
	}
	return false
}

func assertNoSensitive(t *testing.T, rails ...[]Movie) {
	t.Helper()
	for _, rail := range rails {
		for _, m := range rail {
			if IsSensitiveContent(m) {
				t.Fatalf("sensitive title leaked into rails: %s", m.Slug)
			}
		}
	}
}

func TestCorpusSpiderVerse(t *testing.T) {
	pool := corpusPool()
	target := find(pool, "nguoi-nhen-du-hanh-vu-tru-nhen")
	res := Recommend(target, without(pool, target.Slug), UserContext{})
	if !hasSlug(res.SameSeries, "nguoi-nhen-vu-tru-moi") {
		t.Fatalf("same_series missing Into Spider-Verse, got %v", slugs(res.SameSeries))
	}
	if hasSlug(res.SameSeries, "nguoi-nhen-2") {
		t.Fatalf("live Spider-Man must not be same_series for animated target")
	}
	if !hasSlug(res.SimilarContent, "nguoi-nhen-2") && !hasSlug(res.SimilarContent, "nguoi-nhen-3") {
		t.Fatalf("similar_content must prioritize live Spider-Man peers, got %v", slugs(res.SimilarContent))
	}
	assertNoSensitive(t, res.SameSeries, res.SimilarContent, res.YouMayLike)
}

func TestCorpusJujutsuNoLiveUSSeries(t *testing.T) {
	pool := corpusPool()
	target := find(pool, "chu-thuat-hoi-chien-phan-2")
	res := Recommend(target, without(pool, target.Slug), UserContext{})
	if !hasSlug(res.SameSeries, "chu-thuat-hoi-chien-phan-1") {
		t.Fatalf("same_series missing JJK S1, got %v", slugs(res.SameSeries))
	}
	if hasSlug(res.SameSeries, "thi-tran-smallville-phan-4") || hasSlug(res.SimilarContent, "thi-tran-smallville-phan-4") {
		t.Fatalf("Smallville must not appear for JJK")
	}
	assertNoSensitive(t, res.SameSeries, res.SimilarContent, res.YouMayLike)
}

func TestCorpusInvinciblePurity(t *testing.T) {
	pool := corpusPool()
	target := find(pool, "bat-kha-chien-bai-phan-4")
	res := Recommend(target, without(pool, target.Slug), UserContext{})
	for _, bad := range []string{"tinh-nghia-giang-ho", "beo-bat-kha-chien-bai", "tran-tuong-6h30-nguoi-me-quyen-anh"} {
		if hasSlug(res.SameSeries, bad) {
			t.Fatalf("same_series leaked impostor %s: %v", bad, slugs(res.SameSeries))
		}
	}
	if !hasSlug(res.SameSeries, "bat-kha-chien-bai-phan-1") {
		t.Fatalf("same_series missing S1, got %v", slugs(res.SameSeries))
	}
	// "dành cho người lớn" description must not flag seasons as sensitive
	if IsSensitiveContent(find(pool, "bat-kha-chien-bai-phan-1")) {
		t.Fatal("Invincible S1 description must not be treated as adult content")
	}
	assertNoSensitive(t, res.SameSeries, res.SimilarContent, res.YouMayLike)
}

func TestCorpusDemonSlayerNotWukong(t *testing.T) {
	pool := corpusPool()
	target := find(pool, "thanh-guom-diet-quy-phan-1-kamado-tanjiro-lap-chi")
	res := Recommend(target, without(pool, target.Slug), UserContext{})
	if hasSlug(res.SameSeries, "ngo-khong-ky-truyen") {
		t.Fatalf("Immortal Demon Slayer (Wukong) must not be same_series, got %v", slugs(res.SameSeries))
	}
	if !hasSlug(res.SameSeries, "thanh-guom-diet-quy-phan-2-chuyen-tau-vo-tan") {
		t.Fatalf("same_series missing Demon Slayer S2, got %v", slugs(res.SameSeries))
	}
}

func TestCorpusOnePieceLiveNotAnimeSeries(t *testing.T) {
	pool := corpusPool()
	target := find(pool, "dao-hai-tac-live-action-phan-2")
	res := Recommend(target, without(pool, target.Slug), UserContext{})
	if hasSlug(res.SameSeries, "dao-hai-tac-5-loi-nguyen-thanh-kiem") {
		t.Fatalf("anime One Piece movie must not be same_series for live-action target")
	}
}

func TestCorpusGuestRailsExcludeAdult(t *testing.T) {
	pool := corpusPool()
	target := find(pool, "nguoi-nhen-du-hanh-vu-tru-nhen")
	res := Recommend(target, without(pool, target.Slug), UserContext{})
	for _, m := range append(res.YouMayLike, res.SimilarContent...) {
		if strings.Contains(m.Slug, "nghien-sex") || strings.Contains(m.Slug, "cuong-dam") {
			t.Fatalf("adult decoy in rails: %s", m.Slug)
		}
	}
	assertNoSensitive(t, res.YouMayLike)
}

func TestCorpusStrangerThingsTalesNotTrollhunters(t *testing.T) {
	st := Movie{
		Slug: "cau-be-mat-tich-chuyen-nam-85", Name: "Cậu Bé Mất Tích: Chuyện Năm 85",
		OriginName: "Stranger Things: Tales From '85",
		Genres:     []string{"khoa-hoc", "vien-tuong", "bi-an", "hanh-dong", "phieu-luu", "tam-ly", "hoathinh"},
		Country:    "Âu Mỹ",
	}
	troll := Movie{
		Slug: "tho-san-yeu-tinh-truyen-thuyet-arcadia-phan-1",
		Name: "Thợ Săn Yêu Tinh: Truyền Thuyết Arcadia (Phần 1)",
		OriginName: "Trollhunters: Tales of Arcadia (Season 1)",
		Genres:     []string{"hanh-dong", "phieu-luu", "vien-tuong", "hoat-hinh"},
		Country:    "Âu Mỹ",
	}
	stDoc := Movie{
		Slug: "cuoc-phieu-luu-cuoi-hau-truong-cau-be-mat-tich-5",
		Name: "Cuộc phiêu lưu cuối: Hậu trường Cậu bé mất tích 5",
		OriginName: "One Last Adventure: The Making of Stranger Things 5",
		Genres:     []string{"tai-lieu", "hanh-dong", "bi-an"},
		Country:    "Âu Mỹ",
	}
	if ScoreSeriesMatch(st, troll) != 0 {
		t.Fatalf("Trollhunters must not be same_series for Tales From '85, score=%f", ScoreSeriesMatch(st, troll))
	}
	res := Recommend(st, []Movie{troll, stDoc, find(corpusPool(), "nhat-ky-co-nang-nghien-sex")}, UserContext{})
	if hasSlug(res.SameSeries, troll.Slug) {
		t.Fatalf("same_series leaked Trollhunters: %v", slugs(res.SameSeries))
	}
}

func TestCorpusOnePieceLiveActionFranchisePriority(t *testing.T) {
	pool := corpusPool()
	// Noise that previously outranked OP peers under weak commercial matching.
	for _, noise := range []Movie{
		{Slug: "thi-tran-smallville-phan-2", Name: "Thị Trấn Smallville (Phần 2)", OriginName: "Smallville (Season 2)", Genres: []string{"khoa-hoc", "vien-tuong", "hanh-dong", "phieu-luu", "chinh-kich", "tam-ly", "series"}, Country: "Âu Mỹ"},
		{Slug: "doi-dac-nhiem-shield-phan-2", Name: "Đội Đặc Nhiệm SHIELD (Phần 2)", OriginName: "Agents of S.H.I.E.L.D. (Season 2)", Genres: []string{"hanh-dong", "phieu-luu", "khoa-hoc", "vien-tuong", "series"}, Country: "Âu Mỹ"},
		{Slug: "hiep-si-xe-den-phan-1", Name: "Hiệp Sĩ Xe Đen (Phần 1)", OriginName: "Knight Rider (Season 1)", Genres: []string{"hanh-dong", "phieu-luu", "khoa-hoc", "series"}, Country: "Âu Mỹ"},
	} {
		pool = append(pool, noise)
	}
	target := find(pool, "dao-hai-tac-live-action-phan-2")
	opAnime := find(pool, "dao-hai-tac-5-loi-nguyen-thanh-kiem")
	small := find(pool, "thi-tran-smallville-phan-4")
	shield := find(pool, "doi-dac-nhiem-shield-phan-2")
	opScore := ScoreSimilarContent(target, opAnime)
	for _, noise := range []Movie{small, shield, find(pool, "hiep-si-xe-den-phan-1")} {
		ns := ScoreSimilarContent(target, noise)
		if opScore <= ns {
			t.Fatalf("One Piece peer score %f must beat %s (%f)", opScore, noise.Slug, ns)
		}
		if opScore-ns < 3 {
			t.Fatalf("franchise margin too thin: OP %f vs %s %f (need ≥3)", opScore, noise.Slug, ns)
		}
	}
	res := Recommend(target, without(pool, target.Slug), UserContext{})
	if len(res.SimilarContent) == 0 {
		t.Fatal("empty similar_content")
	}
	// First similar row must be a franchise peer (dominates over Smallville/SHIELD/Knight Rider).
	top := res.SimilarContent[0]
	if !strings.Contains(top.Slug, "dao-hai-tac") && !strings.Contains(top.Slug, "one-piece") {
		t.Fatalf("similar[0] must be One Piece franchise, got %s (rail=%v)", top.Slug, slugs(res.SimilarContent))
	}
	if !hasSlug(res.SimilarContent, opAnime.Slug) && !hasAnySlugPrefix(res.SimilarContent, "dao-hai-tac", "one-piece") {
		t.Fatalf("similar_content missing One Piece franchise peers, got %v", slugs(res.SimilarContent))
	}
}

func TestCorpusShortlistKeepsFranchisePeers(t *testing.T) {
	// Large filler pool + sparse invincible peers must still surface seasons.
	target := find(corpusPool(), "bat-kha-chien-bai-phan-4")
	cands := make([]Movie, 0, 500)
	for i := 0; i < 400; i++ {
		cands = append(cands, Movie{
			Slug: "pad-" + strings.Repeat("x", 1) + string(rune('a'+(i%26))) + string(rune('0'+i%10)),
			Name: "Pad", OriginName: "Pad Film", Genres: []string{"hai-huoc"}, Country: "trung-quoc",
		})
	}
	// unique slugs
	for i := range cands {
		cands[i].Slug = "pad-" + itoaCorpus(i)
	}
	cands = append(cands, without(corpusPool(), target.Slug)...)
	res := Recommend(target, cands, UserContext{})
	if !hasSlug(res.SameSeries, "bat-kha-chien-bai-phan-1") {
		t.Fatalf("shortlist dropped true Invincible peer, same=%v", slugs(res.SameSeries))
	}
}

func itoaCorpus(i int) string {
	if i == 0 {
		return "0"
	}
	var b [12]byte
	n := len(b)
	for i > 0 {
		n--
		b[n] = byte('0' + i%10)
		i /= 10
	}
	return string(b[n:])
}
