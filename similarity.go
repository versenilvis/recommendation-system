package recommender

import (
	"strings"
)

func rootOverlap(aRoots, bRoots []string) int {
	best := 0
	for _, ar := range aRoots {
		for _, br := range bRoots {
			if ar == "" || br == "" {
				continue
			}
			if ar == br {
				if !isTooGenericRoot(ar) {
					return 2
				}
				if best < 1 {
					best = 1
				}
			} else {
				if rootPrefixOverlap(ar, br) && len(strings.Fields(ar)) >= 3 && len(strings.Fields(br)) >= 3 {
					if prefix := rootCommonPrefix(ar, br); prefix != "" && !isTooGenericRoot(prefix) {
						return 2
					}
				}
				if prefix := rootCommonPrefix(ar, br); prefix != "" {
					if !isTooGenericRoot(prefix) {
						return 2
					}
				}
				if rootContains(ar, br) || rootContains(br, ar) {
					shorter := ar
					if len(strings.Fields(br)) < len(strings.Fields(ar)) {
						shorter = br
					}
					if !isTooGenericRoot(shorter) {
						return 2
					}
				}
				if best < 1 && (rootContains(ar, br) || rootContains(br, ar) || rootPrefixOverlap(ar, br)) {
					best = 1
				}
			}
		}
	}
	return best
}

func rootPrefixOverlap(a, b string) bool {
	aT := meaningfulTokens(a)
	bT := meaningfulTokens(b)
	common := 0
	for i := 0; i < len(aT) && i < len(bT); i++ {
		if aT[i] == bT[i] {
			common++
		} else {
			break
		}
	}
	return common >= 2
}

func rootCommonPrefix(a, b string) string {
	aT := meaningfulTokens(a)
	bT := meaningfulTokens(b)
	var common []string
	for i := 0; i < len(aT) && i < len(bT); i++ {
		if aT[i] == bT[i] {
			common = append(common, aT[i])
		} else {
			break
		}
	}
	return strings.Join(common, " ")
}

func meaningfulTokens(s string) []string {
	var out []string
	for _, t := range strings.Fields(s) {
		if !isRootStopword(t) {
			out = append(out, t)
		}
	}
	return out
}

func rootContains(a, b string) bool {
	aT := strings.Fields(a)
	bT := strings.Fields(b)
	if len(aT) == 0 || len(bT) == 0 {
		return false
	}
	shorter, longer := bT, aT
	if len(aT) < len(bT) {
		shorter, longer = aT, bT
	}
	if len(shorter) == 1 {
		tok := shorter[0]
		if isTooGenericRoot(tok) {
			return false
		}
		for _, t := range longer {
			if t == tok {
				return true
			}
		}
		return false
	}
	for i := 0; i <= len(longer)-len(shorter); i++ {
		match := true
		for j, tok := range shorter {
			if longer[i+j] != tok {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func SeriesOrderLess(a, b Movie) bool {
	if a.Year != b.Year {
		if a.Year == 0 {
			return false
		}
		if b.Year == 0 {
			return true
		}
		return a.Year < b.Year
	}
	aSer := isSeriesMovie(a)
	bSer := isSeriesMovie(b)
	if aSer != bSer {
		return aSer
	}
	aPart := getPartNumber(a.Name, a.Slug)
	bPart := getPartNumber(b.Name, b.Slug)
	if aPart != bPart {
		return aPart < bPart
	}
	return a.Name < b.Name
}

func sameAnimationProfile(a, b Movie) bool {
	aAnim, aKnown := animationProfile(a)
	bAnim, bKnown := animationProfile(b)
	if !aKnown || !bKnown {
		return false
	}
	return aAnim == bAnim
}

func animationFormat(m Movie) string {
	isAnim := false
	for _, g := range m.Genres {
		gl := strings.ToLower(g)
		if gl == "hoat-hinh" || gl == "anime" || gl == "hoathinh" {
			isAnim = true
			break
		}
	}
	if !isAnim {
		if len(m.Genres) == 0 {
			return ""
		}
		return "live"
	}
	country := strings.ToLower(strings.TrimSpace(m.Country))
	switch country {
	case "nhật bản", "nhat-ban", "japan":
		return "anime"
	case "hàn quốc", "han-quoc", "korea":
		return "anime"
	default:
		return "cartoon"
	}
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

func isSeriesMovie(m Movie) bool {
	for _, g := range m.Genres {
		gl := strings.ToLower(g)
		if gl == "phim-bo" || gl == "tv-shows" {
			return true
		}
	}
	nl := strings.ToLower(m.Name)
	if strings.Contains(nl, "phần ") || strings.Contains(nl, "season ") ||
		strings.Contains(nl, "tập ") || strings.Contains(nl, " bộ") {
		return true
	}
	sl := strings.ToLower(m.Slug)
	return strings.Contains(sl, "-phan-") || strings.Contains(sl, "-season-") ||
		strings.Contains(sl, "-ss-") || strings.Contains(sl, "-tap-")
}
