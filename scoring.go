package recommender

import "strings"

func ScoreSeriesMatch(target, cm Movie) float64 {
	actorOverlap := intersectionCount(cm.Actors, target.Actors)
	directorOverlap := intersectionCount(cm.Directors, target.Directors)

	engMatch, viMatch, matchLevel := getFranchiseMatchLevels(target, cm)

	// Weak franchise signal without shared cast/crew is usually a false positive.
	if matchLevel == 1 && actorOverlap == 0 && directorOverlap == 0 {
		matchLevel = 0
	}

	// Root containment alone is too loose ("Invincible" ⊂ "Invincible Medic").
	// Keep only series-base-compatible pairs or strong distinctive roots (larva…).
	if matchLevel > 0 && !seriesBasesCompatible(target, cm) && !hasStrongSharedRoot(target, cm) {
		matchLevel = 0
	}

	// Animated vs live-action never share same_series.
	if formatsConflict(target, cm) {
		matchLevel = 0
	}

	// Amazon Invincible seasons/specials must not absorb unrelated "Invincible *" titles
	// (e.g. Chinese films with origin_name exactly "Invincible").
	if isAmazonInvincibleShow(target) != isAmazonInvincibleShow(cm) {
		if isAmazonInvincibleShow(target) || isAmazonInvincibleShow(cm) {
			matchLevel = 0
		}
	}

	if matchLevel == 0 && !formatsConflict(target, cm) {
		crewMatch := directorOverlap >= 1 || actorOverlap >= 2
		exactBase := sharedExactSeriesBase(target, cm)

		// Multi-continuity franchises (Spider-Man reboots): cast/crew required.
		if exactBase != "" && isMultiContinuityFranchise(exactBase) {
			if crewMatch {
				matchLevel = 2
			}
		} else if seriesBasesCompatible(target, cm) {
			// True seasons / extensions (Invincible S1–S4, Larva Island).
			if exactBase != "" {
				matchLevel = 2
			} else {
				matchLevel = 2
			}
		} else if crewMatch {
			noRoots := len(franchiseRoots(target)) == 0 && len(franchiseRoots(cm)) == 0
			if noRoots {
				matchLevel = 1
			}
		}
	}

	if matchLevel == 0 {
		return 0
	}

	score := 10.0
	if matchLevel == 2 {
		score += 8.0
	} else {
		score += 3.0
	}

	if engMatch > 0 && viMatch > 0 {
		score += 5.0
	}

	score += float64(actorOverlap) * 2.0
	score += float64(directorOverlap) * 3.0
	score += float64(genreIntersectionCount(cm.Genres, target.Genres)) * 1.0

	if sameAnimationProfile(target, cm) {
		score += 2.0
	}
	if sameCountry(target, cm) {
		score += 3.0
	}

	return score
}

func ScoreSimilarContent(target, cm Movie) float64 {
	genreOverlap := genreIntersectionCount(cm.Genres, target.Genres)
	actorOverlap := intersectionCount(cm.Actors, target.Actors)
	directorOverlap := intersectionCount(cm.Directors, target.Directors)
	engMatch, viMatch, franchiseLv := getFranchiseMatchLevels(target, cm)

	tBase := commercialSeriesBase(target.OriginName)
	cBase := commercialSeriesBase(cm.OriginName)
	sharedCommercialBase := tBase != "" && cBase != "" && tBase == cBase
	// Prefix commercial match: "one piece" vs "one piece curse of the sacred sword"
	if !sharedCommercialBase && tBase != "" && cBase != "" {
		if commercialBaseRelated(tBase, cBase) {
			sharedCommercialBase = true
		}
	}
	tVi := commercialSeriesBase(target.Name)
	cVi := commercialSeriesBase(cm.Name)
	if tVi != "" && cVi != "" && (tVi == cVi || commercialBaseRelated(tVi, cVi)) {
		sharedCommercialBase = true
	}

	if genreOverlap == 0 && actorOverlap == 0 && directorOverlap == 0 &&
		franchiseLv == 0 && !sharedCommercialBase && !sameCountry(target, cm) {
		return 0
	}

	score := 0.0

	switch franchiseLv {
	case 2:
		score += 8.0
	case 1:
		score += 4.0
	}

	if engMatch > 0 && viMatch > 0 {
		score += 3.0
	}

	// Shared commercial series base (e.g. spider-man, one piece LA↔anime) should
	// dominate "Nội dung tương tự" over merely same-genre live-action series.
	compatibleSeries := seriesBasesCompatible(target, cm)
	if sharedCommercialBase || compatibleSeries {
		if franchiseLv == 0 {
			score += 22.0
		} else {
			score += 10.0
		}
	}

	// Hard separate anime/cartoon from live-action unless same franchise / series name.
	if formatsConflict(target, cm) {
		if franchiseLv == 0 && !sharedCommercialBase && !compatibleSeries {
			score -= 15.0
		} else {
			// Soft demotion for live remakes / alternate formats of the same franchise.
			// Keep well below the commercial boost so franchise peers still rank first.
			score -= 2.0
		}
	} else if sameAnimationProfile(target, cm) {
		tFmt := animationFormat(target)
		cFmt := animationFormat(cm)
		if tFmt != "" && tFmt == cFmt {
			score += 4.0
		}
	}

	if sameCountry(target, cm) {
		score += 4.0
	}

	if isSeriesMovie(target) == isSeriesMovie(cm) {
		score += 1.0
	}

	score += float64(genreOverlap) * 2.0
	score += float64(actorOverlap) * 1.5
	score += float64(directorOverlap) * 2.0

	// Keep adult titles out of non-adult neighborhoods.
	if IsSensitiveContent(cm) && !IsSensitiveContent(target) {
		score -= 20.0
	}

	return score
}

func ScorePersonalised(
	cm Movie,
	userGenreScores map[string]float64,
	coWatchedMap map[string]bool,
	recentGenresMap map[string]int,
) float64 {
	score := 0.0

	for _, g := range cm.Genres {
		ng := normalizeGenre(g)
		if s, ok := userGenreScores[g]; ok {
			score += s
		} else if s, ok := userGenreScores[ng]; ok {
			score += s
		}
	}

	if coWatchedMap[cm.Slug] {
		score += 5.0
	}

	for _, g := range cm.Genres {
		ng := normalizeGenre(g)
		if count, ok := recentGenresMap[g]; ok {
			score += float64(count) * 1.5
		} else if count, ok := recentGenresMap[ng]; ok {
			score += float64(count) * 1.5
		}
	}

	return score
}

// ScoreYouMayLike blends personalisation with soft similarity so guests
// (personalisation = 0) still get coherent, non-random suggestions.
func ScoreYouMayLike(target, cm Movie, ctx UserContext) float64 {
	if IsSensitiveContent(cm) && !IsSensitiveContent(target) {
		return -100
	}

	score := ScorePersonalised(cm, ctx.GenreScores, ctx.CoWatchedMovies, ctx.RecentGenres)

	sim := ScoreSimilarContent(target, cm)
	if sim > 0 {
		score += sim * 0.45
	}

	if sameCountry(target, cm) {
		score += 3.0
	}

	// Prefer same animation format for discovery rail.
	if sameAnimationProfile(target, cm) {
		score += 2.0
	} else if formatsConflict(target, cm) {
		score -= 12.0
	}

	return score
}

func getFranchiseMatchLevels(target, cm Movie) (engMatch, viMatch, franchiseLv int) {
	engMatch = rootOverlap(franchiseRoots(target), franchiseRoots(cm))
	viMatch = rootOverlap(vietnameseRoots(target), vietnameseRoots(cm))
	franchiseLv = engMatch
	if viMatch > franchiseLv {
		franchiseLv = viMatch
	}
	return
}

// commercialSeriesBase is seriesBase folded for cross-format franchise matching
// (live-action vs anime One Piece, etc.).
func commercialSeriesBase(name string) string {
	return normRoot(seriesBase(name))
}

// commercialBaseRelated is true when one commercial base is a distinctive prefix
// of the other (one piece ⊂ one piece curse sacred sword) without weak tokens.
func commercialBaseRelated(a, b string) bool {
	if a == "" || b == "" || a == b {
		return a != "" && a == b
	}
	aT, bT := strings.Fields(a), strings.Fields(b)
	if len(aT) == 0 || len(bT) == 0 {
		return false
	}
	shorter, longer := aT, bT
	if len(aT) > len(bT) {
		shorter, longer = bT, aT
	}
	// Require at least 2 tokens for prefix relation (avoid "one" matching everything)
	// OR a single distinctive token of length >= 8.
	if len(shorter) == 1 {
		if isWeakContainmentToken(shorter[0]) || isTooGenericRoot(shorter[0]) {
			return false
		}
		if len([]rune(shorter[0])) < 8 {
			return false
		}
	}
	if !isTokenPrefix(shorter, longer) {
		return false
	}
	// Reject if the shared head is only weak media words
	head := strings.Join(shorter, " ")
	if isTooGenericRoot(head) || isWeakContainmentToken(head) {
		return false
	}
	return true
}
