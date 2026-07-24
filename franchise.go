package recommender

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

func extractRoots(name string) []string {
	seen := make(map[string]bool)
	var roots []string
	add := func(s string) {
		s = normRoot(s)
		if s == "" || seen[s] || isTooGenericRoot(s) {
			return
		}
		seen[s] = true
		roots = append(roots, s)
	}
	add(seriesBase(name))
	add(seriesSubtitle(name))
	return roots
}

func franchiseRoots(m Movie) []string {
	return extractRoots(m.OriginName)
}

func vietnameseRoots(m Movie) []string {
	return extractRoots(m.Name)
}

func SeriesBase(name string) string {
	return seriesBase(name)
}

// franchiseExtensionTokens are allowed extra tokens when one series base is a
// prefix of another (e.g. "larva" ↔ "larva island"). Arbitrary second titles
// like "invincible medic" / "invincible fatty" must NOT match.
var franchiseExtensionTokens = map[string]bool{
	"island": true, "movie": true, "film": true, "zero": true, "ova": true,
	"special": true, "returns": true, "begins": true, "legend": true,
	"chronicles": true, "tales": true, "story": true, "shippuden": true,
	"kai": true, "z": true, "gt": true, "super": true, "brotherhood": true,
	"the": true, "of": true, "and": true, "vs": true, "versus": true,
	"part": true, "chapter": true, "book": true, "final": true, "ultimate": true,
	"complete": true, "collection": true, "remastered": true, "edition": true,
	"black": true, "white": true, // e.g. sword art online progressive
	"atom": true, "eve": true, "presenting": true, // Invincible specials
}

// seriesBasesCompatible reports whether two titles share a real series identity
// (exact base or base + franchise extension), not merely a shared adjective/token.
func seriesBasesCompatible(a, b Movie) bool {
	candidates := [][2]string{
		{a.OriginName, b.OriginName},
		{a.Name, b.Name},
		{a.OriginName, b.Name},
		{a.Name, b.OriginName},
		{slugBase(a.Slug), slugBase(b.Slug)},
	}
	for _, pair := range candidates {
		if basesCompatible(pair[0], pair[1]) {
			return true
		}
	}
	return false
}

func basesCompatible(a, b string) bool {
	a = normRoot(seriesBase(a))
	b = normRoot(seriesBase(b))
	if a == "" || b == "" {
		return false
	}
	// Exact series-base equality is always a real match (incl. "spider man", "invincible"),
	// even when those tokens are treated as too-generic for fuzzy root containment.
	if a == b {
		return true
	}
	aT := strings.Fields(a)
	bT := strings.Fields(b)
	if len(aT) == 0 || len(bT) == 0 {
		return false
	}
	if isTokenPrefix(aT, bT) {
		return extensionTokensOK(bT[len(aT):]) && !isTooGenericRoot(strings.Join(aT, " "))
	}
	if isTokenPrefix(bT, aT) {
		return extensionTokensOK(aT[len(bT):]) && !isTooGenericRoot(strings.Join(bT, " "))
	}
	return false
}

func isTokenPrefix(prefix, full []string) bool {
	if len(prefix) == 0 || len(prefix) > len(full) {
		return false
	}
	for i, t := range prefix {
		if full[i] != t {
			return false
		}
	}
	return true
}

func extensionTokensOK(extra []string) bool {
	if len(extra) == 0 {
		return true
	}
	for _, t := range extra {
		if franchiseExtensionTokens[t] {
			continue
		}
		// pure numbers already stripped by seriesBase; reject unknown words
		return false
	}
	return true
}

// multiContinuityFranchise bases have many unrelated reboots; same_series needs cast/crew.
func isMultiContinuityFranchise(base string) bool {
	switch normRoot(base) {
	case "spider man", "spiderman", "batman", "superman", "iron man", "ironman",
		"avengers", "x men", "xmen", "captain america", "wonder woman":
		return true
	}
	return false
}

func sharedExactSeriesBase(a, b Movie) string {
	pairs := [][2]string{
		{normRoot(seriesBase(a.OriginName)), normRoot(seriesBase(b.OriginName))},
		{normRoot(seriesBase(a.Name)), normRoot(seriesBase(b.Name))},
	}
	for _, p := range pairs {
		if p[0] != "" && p[0] == p[1] {
			return p[0]
		}
	}
	return ""
}

// hasStrongSharedRoot keeps true spin-offs (larva ↔ larva island) while rejecting
// weak single-token containment (invincible ⊂ invincible medic).
func hasStrongSharedRoot(a, b Movie) bool {
	check := func(aRoots, bRoots []string) bool {
		for _, ar := range aRoots {
			for _, br := range bRoots {
				if ar == "" || br == "" {
					continue
				}
				if ar == br {
					if !isTooGenericRoot(ar) && !isWeakContainmentToken(ar) {
						return true
					}
					continue
				}
				// Shared prefix: multi-token always; single distinctive token (larva) ok.
				if prefix := rootCommonPrefix(ar, br); prefix != "" {
					n := len(strings.Fields(prefix))
					if n >= 2 && !isTooGenericRoot(prefix) {
						return true
					}
					if n == 1 && !isTooGenericRoot(prefix) && !isWeakContainmentToken(prefix) {
						return true
					}
				}
				// single-token containment only for distinctive tokens (larva, bleach…)
				if rootContains(ar, br) || rootContains(br, ar) {
					shorter := ar
					if len(strings.Fields(br)) < len(strings.Fields(ar)) {
						shorter = br
					}
					toks := strings.Fields(shorter)
					if len(toks) == 1 && !isTooGenericRoot(shorter) && !isWeakContainmentToken(shorter) {
						return true
					}
					if len(toks) >= 2 && !isTooGenericRoot(shorter) {
						return true
					}
				}
			}
		}
		return false
	}
	return check(franchiseRoots(a), franchiseRoots(b)) || check(vietnameseRoots(a), vietnameseRoots(b))
}

func isWeakContainmentToken(root string) bool {
	switch normRoot(root) {
	case "invincible", "spider", "batman", "superman", "hero", "king", "legend",
		"dragon", "warrior", "master", "love", "war", "girl", "boy", "man", "woman",
		"nhen", "nhện":
		return true
	}
	return false
}

func seriesBase(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	if s == "" {
		return ""
	}
	if idx := strings.Index(s, ":"); idx > 2 {
		s = s[:idx]
	}
	for _, marker := range []string{
		" phần ", " - phần ", "(phần ",
		" season ", " - season ", "(season ", ": season ",
		" ss", " (ss",
		" (movie)", " - movie",
		" movie",
		" part ", " - part ",
	} {
		if idx := strings.Index(s, marker); idx > 0 {
			s = s[:idx]
			break
		}
	}
	return trimTrailingNumber(strings.TrimSpace(s))
}

func seriesSubtitle(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	idx := strings.Index(s, ":")
	if idx <= 2 {
		return ""
	}
	subtitle := strings.TrimSpace(s[idx+1:])
	for _, marker := range []string{" phan ", " season ", "(season", "(phan "} {
		if midx := strings.Index(subtitle, marker); midx > 0 {
			subtitle = subtitle[:midx]
			break
		}
	}
	return trimTrailingNumber(strings.TrimSpace(subtitle))
}

func slugBase(slug string) string {
	slug = strings.ToLower(slug)
	parts := strings.Split(slug, "-")
	var keep []string
	skip := false
	for _, p := range parts {
		if skip {
			skip = false
			continue
		}
		switch p {
		case "phan", "season", "ss", "part", "tap", "movie":
			skip = true
			continue
		}
		if _, err := strconv.Atoi(p); err == nil {
			continue
		}
		keep = append(keep, p)
	}
	return strings.Join(keep, " ")
}

func normRoot(s string) string {
	s = strings.ToLower(s)
	r := strings.NewReplacer(":", " ", "-", " ", "_", " ", "(", " ", ")", " ", ".", " ", ",", " ")
	s = r.Replace(s)
	tokens := strings.Fields(s)
	var keep []string
	for _, t := range tokens {
		if !isRootStopword(t) {
			keep = append(keep, t)
		}
	}
	return strings.Join(keep, " ")
}

func isRootStopword(t string) bool {
	if utf8.RuneCountInString(t) <= 1 {
		return true
	}
	if _, err := strconv.Atoi(t); err == nil {
		return true
	}
	switch t {
	case "the", "a", "an", "of", "and", "or",
		"phim", "movie", "film",
		"phan", "season", "ss", "tap", "part",
		"ban", "truyen", "nguyen", "tac",
		"chu", "chú", "co", "cô", "cau", "cậu", "be", "bé", "chang", "chàng", "nang", "nàng",
		"nhung", "những", "cac", "các", "mot", "một", "su", "sự", "ke", "kẻ", "nguoi", "người",
		"thuyet", "minh", "long", "tieng", "vietsub",
		"into", "across", "no", "way", "far", "from", "home",
		"beyond", "through", "within", "without", "along",
		"in", "on", "at", "to", "by", "for", "with",
		"khong", "không", "có", "va", "và", "hoac", "hoặc",
		"nhưng", "cua", "của", "cho", "voi", "với",
		"tai", "tại", "trong", "ngoai", "ngoài", "tren", "trên",
		"duoi", "dưới", "den", "đến", "di", "đi", "ve", "về",
		"nhu", "như", "de", "để", "boi", "bởi", "tu", "từ",
		"o", "ở":
		return true
	}
	return false
}

func isTooGenericRoot(root string) bool {
	tokens := strings.Fields(root)
	if len(tokens) == 0 {
		return true
	}
	meaningful := 0
	var firstToken string
	for _, t := range tokens {
		if !isRootStopword(t) {
			meaningful++
			if firstToken == "" {
				firstToken = t
			}
		}
	}
	if meaningful == 0 {
		return true
	}
	if meaningful == 1 {
		// Very short leftovers after stopword stripping are weak false positives
		// (e.g. "người nhện" → "nhện"). Keep 5+ letter specific tokens like "larva".
		if utf8.RuneCountInString(firstToken) <= 4 {
			return true
		}
		switch firstToken {
		case "love", "hero", "fight", "dark", "blue", "dead", "kill", "fire", "gold", "time",
			"star", "moon", "king", "lord", "wife", "game", "show", "play", "team", "crew",
			"club", "gang", "city", "town", "home", "road", "gate", "wall", "path", "land",
			"wind", "cold", "heat", "warm", "wild", "free", "best", "last", "first", "next",
			"past", "girl", "lady", "host", "boss", "work", "life", "soul", "mind", "heart",
			"body", "face", "hand", "foot", "baby", "child", "kids", "born", "live", "hope",
			"fear", "hate", "wish", "dream", "real", "fake", "true", "liar", "deal", "cash",
			"money", "rich", "poor", "hard", "easy", "safe", "risk", "hell", "heaven", "devil",
			"angel", "ghost", "witch", "beast", "magic", "power", "force", "truth", "secret",
			"death", "sleep", "wake", "lost", "found", "black", "white", "green", "yellow",
			"spider", "batman", "superman", "nhen", "nhện":
			return true
		}
		return false
	}
	if meaningful >= 3 {
		return false
	}
	switch root {
	case
		"love", "tinh yeu", "tình yêu",
		"revenge", "bao thu", "báo thù", "tra thu", "trả thù",
		"hero", "fight", "war",
		"dac vu", "đặc vụ", "mat vu", "mật vụ",
		"canh sat", "cảnh sát", "canh binh", "cảnh binh",
		"nguoi nhen", "người nhện",
		"spider man", "spiderman",
		"batman", "superman", "iron man",
		"sieu nhan", "siêu nhân",
		"sieu anh hung", "siêu anh hùng",
		"avengers", "marvel", "dc":
		return true
	}
	return false
}

func trimTrailingNumber(s string) string {
	for {
		parts := strings.Fields(s)
		if len(parts) <= 1 {
			break
		}
		last := parts[len(parts)-1]
		if _, err := strconv.Atoi(last); err == nil {
			s = strings.Join(parts[:len(parts)-1], " ")
			continue
		}
		switch last {
		case "z", "gt", "kai", "super", "shippuden", "brotherhood":
			s = strings.Join(parts[:len(parts)-1], " ")
			continue
		}
		break
	}
	return s
}

func getPartNumber(name string, slug string) int {
	name = strings.ToLower(name)
	slug = strings.ToLower(slug)
	findNumAfter := func(s, prefix string) int {
		idx := strings.Index(s, prefix)
		if idx == -1 {
			return -1
		}
		start := idx + len(prefix)
		end := start
		for end < len(s) && s[end] >= '0' && s[end] <= '9' {
			end++
		}
		if start == end {
			return -1
		}
		if num, err := strconv.Atoi(s[start:end]); err == nil {
			return num
		}
		return -1
	}
	for _, p := range []string{"phần ", "season ", "ss"} {
		if n := findNumAfter(name, p); n != -1 {
			return n
		}
	}
	for _, p := range []string{"-phan-", "-season-", "-ss-"} {
		if n := findNumAfter(slug, p); n != -1 {
			return n
		}
	}
	parts := strings.Split(strings.Trim(slug, "-"), "-")
	if len(parts) > 1 {
		if n, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
			return n
		}
	}
	return 0
}
