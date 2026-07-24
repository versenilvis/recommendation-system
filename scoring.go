package recommender

func ScoreSeriesMatch(target, cm Movie) float64 {
	actorOverlap := intersectionCount(cm.Actors, target.Actors)
	directorOverlap := intersectionCount(cm.Directors, target.Directors)

	engMatch, viMatch, matchLevel := getFranchiseMatchLevels(target, cm)

	// Weak franchise signal without shared cast/crew is usually a false positive
	// (e.g. generic single-token leftovers).
	if matchLevel == 1 && actorOverlap == 0 && directorOverlap == 0 {
		matchLevel = 0
	}

	// Animated vs live-action must never share same_series unless hard crew proof
	// of a true shared production (extremely rare — keep out by default).
	if formatsConflict(target, cm) {
		matchLevel = 0
	}

	if matchLevel == 0 {
		noRoots := len(franchiseRoots(target)) == 0 && len(franchiseRoots(cm)) == 0
		crewMatch := directorOverlap >= 1 || actorOverlap >= 2
		if noRoots && crewMatch && !formatsConflict(target, cm) {
			matchLevel = 1
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

	tBase := normRoot(seriesBase(target.OriginName))
	cBase := normRoot(seriesBase(cm.OriginName))
	sharedCommercialBase := tBase != "" && cBase != "" && tBase == cBase
	tVi := normRoot(seriesBase(target.Name))
	cVi := normRoot(seriesBase(cm.Name))
	if tVi != "" && cVi != "" && tVi == cVi {
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

	// Shared commercial series base (e.g. spider-man) even when roots are "generic"
	// for same_series purposes — still a real franchise signal for similar_content.
	if sharedCommercialBase && franchiseLv == 0 {
		score += 5.0
	}

	// Hard separate anime/cartoon from live-action unless same franchise / series name.
	if formatsConflict(target, cm) {
		if franchiseLv == 0 && !sharedCommercialBase {
			score -= 15.0
		} else {
			// Soft demotion for live remakes / alternate formats of the same franchise.
			score -= 6.0
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
