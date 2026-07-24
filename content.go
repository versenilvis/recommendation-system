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
// Title/slug signals are primary. Description text only matches strongly sexual phrases —
// bare "người lớn" is common for mature animated shows (e.g. Invincible) and must not flag.
func IsSensitiveContent(m Movie) bool {
	for g := range genreSet(m.Genres) {
		switch g {
		case "phim-18", "tinh-duc", "phim-nguoi-lon", "erotic", "adult":
			return true
		}
	}
	// Name + slug only for short sexual keywords (avoid description false positives).
	titleBlob := strings.ToLower(removeDiacritics(m.Name + " " + m.OriginName + " " + m.Slug))
	for _, kw := range []string{
		"sex", "xxx", "18+", "phim-18", "phim 18",
		"cuong dam", "cuong-dam", "nghien sex", "nghien-sex",
		"erotic", "porn", "phim nguoi lon", "phim-nguoi-lon",
		"tinh duc", "tinh-duc", "dam duc", "loan luan",
		"khieu dam", "softcore", "hardcore",
	} {
		if strings.Contains(titleBlob, removeDiacritics(kw)) {
			return true
		}
	}
	// Description: only explicit sexual marketing phrases, not "dành cho người lớn".
	if m.Content != "" {
		desc := strings.ToLower(removeDiacritics(m.Content))
		for _, kw := range []string{
			"phim sex", "phim 18+", "phim nguoi lon", "noi dung 18+",
			"canh nong", "canh sex", "phim cap ba", "phim cap 3",
		} {
			if strings.Contains(desc, removeDiacritics(kw)) {
				return true
			}
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
	"anime", "jujutsu", "chu thuat", "chu-thuat", "naruto",
	// "one piece" alone also matches live-action; handled via live-action override above
	"one piece", "one-piece", "dao hai tac",
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
	"larva", "au trung", "arcane", "avatar the last", "legend of korra",
	"my adventures with superman", "x-men 97", "x men 97",
}

// isAmazonInvincibleShow detects the animated Invincible series/specials even when
// the catalog omits hoat-hinh (common for bat-kha-chien-bai seasons).
// Bare origin "Invincible" alone is ambiguous (many unrelated films use it) — require
// season/special markers, the Vietnamese slug family, or Atom Eve branding.
func isAmazonInvincibleShow(m Movie) bool {
	blob := contentBlob(m)
	// Reject unrelated Chinese/other titles that only share the word "invincible".
	if strings.Contains(blob, "medic") || strings.Contains(blob, "fatty") ||
		strings.Contains(blob, "dragon") || strings.Contains(blob, "commission") ||
		strings.Contains(blob, "constable") || strings.Contains(blob, "swordsman") ||
		strings.Contains(blob, "iron man") || strings.Contains(blob, "wenger") ||
		strings.Contains(blob, "quyen anh") || strings.Contains(blob, "quyền anh") {
		return false
	}
	slug := strings.ToLower(m.Slug)
	if strings.HasPrefix(slug, "bat-kha-chien-bai") {
		return true
	}
	if strings.Contains(blob, "atom eve") || strings.Contains(slug, "atom-eve") {
		if strings.Contains(blob, "invincible") || strings.Contains(blob, "bat kha chien bai") {
			return true
		}
	}

	ob := strings.ToLower(removeDiacritics(m.OriginName))
	if strings.HasPrefix(ob, "invincible") {
		rest := strings.TrimSpace(strings.TrimPrefix(ob, "invincible"))
		// Require explicit season / special — NOT bare "Invincible".
		if strings.Contains(rest, "season") || strings.Contains(rest, "atom") ||
			strings.Contains(rest, "presenting") || strings.Contains(rest, "eve") {
			return true
		}
	}

	name := strings.ToLower(removeDiacritics(m.Name))
	if strings.Contains(name, "bat kha chien bai") &&
		(strings.Contains(name, "phan") || strings.Contains(name, "atom") ||
			strings.Contains(name, "tap dac biet")) {
		return true
	}
	return false
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

	// Explicit live-action branding wins over shared franchise names (One Piece LA vs anime).
	if strings.Contains(blob, "live action") || strings.Contains(blob, "live-action") ||
		strings.Contains(blob, "nguoi dong") || strings.Contains(blob, "người đóng") {
		return "live"
	}

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

	if isAmazonInvincibleShow(m) {
		return "cartoon"
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
