package recommender

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// NormCountry folds display names, slugs, and diacritics into a stable country key.
// e.g. "Nhật Bản", "nhat-ban", "japan" → "nhat-ban"
func NormCountry(c string) string {
	c = strings.ToLower(strings.TrimSpace(c))
	if c == "" {
		return ""
	}
	c = removeDiacritics(c)
	c = strings.NewReplacer("_", " ", "-", " ", "/", " ").Replace(c)
	c = strings.Join(strings.Fields(c), " ")

	switch c {
	case "nhat ban", "japan", "jp", "japanese", "nhat":
		return "nhat-ban"
	case "han quoc", "korea", "kr", "south korea", "hanguk", "han":
		return "han-quoc"
	case "trung quoc", "china", "cn", "chinese", "trung":
		return "trung-quoc"
	case "au my", "my", "usa", "us", "america", "american", "hollywood my", "hollywood-my",
		"united states", "hoa ky", "hoa ki":
		return "au-my"
	case "hong kong", "hk", "hongkong":
		return "hong-kong"
	case "dai loan", "taiwan", "tw":
		return "dai-loan"
	case "thai lan", "thailand", "th":
		return "thai-lan"
	case "viet nam", "vietnam", "vn", "viet":
		return "viet-nam"
	case "anh", "uk", "england", "britain", "great britain", "united kingdom":
		return "anh"
	case "phap", "france", "fr":
		return "phap"
	case "an do", "india", "in":
		return "an-do"
	case "duc", "germany", "de":
		return "duc"
	case "y", "italy", "it":
		return "y"
	case "nga", "russia", "ru":
		return "nga"
	case "canada", "ca":
		return "canada"
	case "malaysia", "mya":
		return "malaysia"
	case "tay ban nha", "spain", "es":
		return "tay-ban-nha"
	}

	return strings.ReplaceAll(c, " ", "-")
}

func sameCountry(a, b Movie) bool {
	ca, cb := NormCountry(a.Country), NormCountry(b.Country)
	return ca != "" && ca == cb
}

func normalizeGenre(g string) string {
	g = strings.ToLower(strings.TrimSpace(g))
	g = removeDiacritics(g)
	g = strings.NewReplacer(" ", "-", "_", "-").Replace(g)
	g = strings.Join(strings.FieldsFunc(g, func(r rune) bool {
		return r == '-' || r == ' '
	}), "-")
	// common aliases
	switch g {
	case "hoathinh", "animation", "cartoon":
		return "hoat-hinh"
	case "tvshows", "tv-show", "series":
		return "phim-bo"
	case "single", "movie":
		return "phim-le"
	case "18", "18plus", "phim18", "nguoi-lon", "adult":
		return "phim-18"
	case "tinhduc", "sex", "erotic":
		return "tinh-duc"
	}
	return g
}

func genreSet(genres []string) map[string]bool {
	m := make(map[string]bool, len(genres))
	for _, g := range genres {
		if n := normalizeGenre(g); n != "" {
			m[n] = true
		}
	}
	return m
}

func genreIntersectionCount(a, b []string) int {
	as := genreSet(a)
	count := 0
	for g := range genreSet(b) {
		if as[g] {
			count++
		}
	}
	return count
}

// IsSensitiveContent flags adult / softcore titles by genre slug or name/slug keywords.
func IsSensitiveContent(m Movie) bool {
	for g := range genreSet(m.Genres) {
		switch g {
		case "phim-18", "tinh-duc", "phim-nguoi-lon", "erotic", "adult":
			return true
		}
	}
	blob := strings.ToLower(removeDiacritics(
		m.Name + " " + m.OriginName + " " + m.Slug + " " + m.Content,
	))
	// keep keywords ASCII-folded so "Cuồng Dâm" matches "cuong dam"
	for _, kw := range []string{
		"sex", "xxx", "18+", "phim-18", "phim 18",
		"cuong dam", "cuong-dam", "nghien sex", "nghien-sex",
		"erotic", "porn", "nguoi lon", "nguoi-lon",
		"tinh duc", "tinh-duc", "dam duc", "loan luan",
		"khiêu dâm", "khieu dam", "softcore", "hardcore",
	} {
		if strings.Contains(blob, removeDiacritics(kw)) {
			return true
		}
	}
	return false
}

func removeDiacritics(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range norm.NFD.String(s) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func contentBlob(m Movie) string {
	return strings.ToLower(removeDiacritics(
		m.Name + " " + m.OriginName + " " + m.Slug,
	))
}

func hasAnimGenre(m Movie) bool {
	for g := range genreSet(m.Genres) {
		switch g {
		case "hoat-hinh", "anime", "animation", "cartoon":
			return true
		}
	}
	return false
}

// anime / cartoon franchise keywords (ASCII-folded, matched as substrings).
var animeKeywords = []string{
	"anime", "jujutsu", "chu thuat", "chu-thuat", "naruto", "one piece", "one-piece",
	"dragon ball", "dragonball", "demon slayer", "thanh guom", "thanh-guom", "kimetsu",
	"bleach", "chainsaw", "attack on titan", "shingeki", "my hero academia",
	"boku no hero", "spy x family", "spy-family", "tokyo ghoul", "death note",
	"fullmetal", "hunter x hunter", "hunterxhunter", "sword art online",
	"one punch", "mob psycho", "jojo", "evangelion", "ghibli", "doraemon",
	"conan", "pokemon", "digimon", "sailor moon", "haikyuu", "blue lock",
	"oshi no ko", "frieren", "solo leveling", "re:zero", "re zero", "re-zero",
	"overlord", "isekai", "shonen", "shounen", "manga", "boruto", "black clover",
	"jujutsu kaisen", "vinland", "hells paradise", "chainsaw man", "reze",
	"studio ghibli", "makoto shinkai", "your name", "suzume", "weathering with you",
}

var cartoonKeywords = []string{
	"spider-verse", "spiderverse", "spider verse", "pixar", "dreamworks",
	"illumination", "cartoon", "hoat hinh", "hoat-hinh", "animated",
	"despicable", "minions", "frozen", "zootopia", "encanto", "moana",
	"toy story", "incredibles", "kung fu panda", "how to train your dragon",
	"larva", "au trung",
}

func hasAnyKeyword(blob string, kws []string) bool {
	for _, kw := range kws {
		if strings.Contains(blob, removeDiacritics(kw)) {
			return true
		}
	}
	return false
}

// animationFormat returns "anime", "cartoon", "live", or "" (unknown).
// Uses genre, country, and title keywords so DB rows missing hoat-hinh still classify.
func animationFormat(m Movie) string {
	blob := contentBlob(m)
	country := NormCountry(m.Country)
	animGenre := hasAnimGenre(m)
	animeKW := hasAnyKeyword(blob, animeKeywords)
	cartoonKW := hasAnyKeyword(blob, cartoonKeywords)

	// Strong anime signals: JP/KR + (genre or keyword), or explicit anime keyword.
	if animeKW || (animGenre && (country == "nhat-ban" || country == "han-quoc")) {
		if country == "nhat-ban" || country == "han-quoc" || animeKW {
			return "anime"
		}
	}

	// Japan + series TV without live-action cast tags and with fantasy/action stack is often anime
	// when title matches known anime OR origin has typical anime romanization patterns.
	if country == "nhat-ban" && (animGenre || animeKW) {
		return "anime"
	}
	// Keyword "jujutsu" etc. even if country missing
	if animeKW {
		return "anime"
	}

	if animGenre || cartoonKW {
		if country == "nhat-ban" || country == "han-quoc" {
			return "anime"
		}
		return "cartoon"
	}

	// Japan titles that look like anime seasons (phần/season in name) + common anime genre stack
	// without explicit live-action markers: treat as anime when origin is romanized JP-style.
	if country == "nhat-ban" && looksLikeAnimeSeries(m) {
		return "anime"
	}

	if len(m.Genres) == 0 && m.OriginName == "" && m.Name == "" {
		return ""
	}
	// If we have any identity signal, default to live-action.
	if len(m.Genres) > 0 || m.OriginName != "" || m.Name != "" {
		return "live"
	}
	return ""
}

func looksLikeAnimeSeries(m Movie) bool {
	// Prefer not to over-classify JP live-action: require season marker or slug pattern
	// plus sci-fi/fantasy-ish genres common for shonen anime dumps in this catalog.
	nl := strings.ToLower(m.Name + " " + m.Slug + " " + m.OriginName)
	season := strings.Contains(nl, "phần") || strings.Contains(nl, "phan") ||
		strings.Contains(nl, "season") || strings.Contains(nl, "-ss-")
	if !season {
		return false
	}
	gs := genreSet(m.Genres)
	fantasy := gs["vien-tuong"] || gs["khoa-hoc"] || gs["phieu-luu"] || gs["hanh-dong"]
	return fantasy
}

func animationProfile(m Movie) (animated bool, known bool) {
	f := animationFormat(m)
	switch f {
	case "anime", "cartoon":
		return true, true
	case "live":
		return false, true
	default:
		return false, false
	}
}

func sameAnimationProfile(a, b Movie) bool {
	aAnim, aKnown := animationProfile(a)
	bAnim, bKnown := animationProfile(b)
	if !aKnown || !bKnown {
		return false
	}
	return aAnim == bAnim
}

// formatsConflict reports anime/cartoon vs live-action mismatch (or anime vs western cartoon).
func formatsConflict(a, b Movie) bool {
	fa, fb := animationFormat(a), animationFormat(b)
	if fa == "" || fb == "" {
		return false
	}
	if fa == fb {
		return false
	}
	// anime vs cartoon is a soft difference (both animated) — not a hard conflict
	aAnim := fa == "anime" || fa == "cartoon"
	bAnim := fb == "anime" || fb == "cartoon"
	if aAnim && bAnim {
		return false
	}
	return true
}
