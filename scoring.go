package recommender

func ScoreSeriesMatch(target, cm Movie) float64 {
	actorOverlap := intersectionCount(cm.Actors, target.Actors)
	directorOverlap := intersectionCount(cm.Directors, target.Directors)

	engMatch, viMatch, matchLevel := getFranchiseMatchLevels(target, cm)

	if matchLevel == 1 && actorOverlap == 0 && directorOverlap == 0 {
		matchLevel = 0
	}

	if matchLevel == 0 {
		noRoots := len(franchiseRoots(target)) == 0 && len(franchiseRoots(cm)) == 0
		crewMatch := directorOverlap >= 1 || actorOverlap >= 2
		if noRoots && crewMatch {
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
	score += float64(intersectionCount(cm.Genres, target.Genres)) * 1.0

	if sameAnimationProfile(target, cm) {
		score += 2.0
	}

	return score
}

func ScoreSimilarContent(target, cm Movie) float64 {
	genreOverlap := intersectionCount(cm.Genres, target.Genres)
	actorOverlap := intersectionCount(cm.Actors, target.Actors)
	directorOverlap := intersectionCount(cm.Directors, target.Directors)

	if genreOverlap == 0 && actorOverlap == 0 && directorOverlap == 0 {
		return 0
	}

	score := 0.0

	engMatch, viMatch, franchiseLv := getFranchiseMatchLevels(target, cm)

	switch franchiseLv {
	case 2:
		score += 8.0
	case 1:
		score += 4.0
	default:
		tBase := normRoot(seriesBase(target.OriginName))
		cBase := normRoot(seriesBase(cm.OriginName))
		if tBase != "" && cBase != "" && tBase == cBase {
			score += 6.0
		}
	}

	if engMatch > 0 && viMatch > 0 {
		score += 3.0
	}

	tFmt := animationFormat(target)
	cFmt := animationFormat(cm)
	if tFmt != "" && cFmt != "" {
		if tFmt == cFmt {
			score += 4.0
		} else if franchiseLv == 0 {
			tBase := normRoot(seriesBase(target.OriginName))
			cBase := normRoot(seriesBase(cm.OriginName))
			if tBase == "" || cBase == "" || tBase != cBase {
				score -= 10.0
			}
		}
	}

	if isSeriesMovie(target) == isSeriesMovie(cm) {
		score += 1.0
	}

	score += float64(genreOverlap) * 2.0
	score += float64(actorOverlap) * 1.5
	score += float64(directorOverlap) * 2.0

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
		if s, ok := userGenreScores[g]; ok {
			score += s
		}
	}

	if coWatchedMap[cm.Slug] {
		score += 5.0
	}

	for _, g := range cm.Genres {
		if count, ok := recentGenresMap[g]; ok {
			score += float64(count) * 1.5
		}
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
