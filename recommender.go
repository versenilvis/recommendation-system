package recommender

import (
	"sort"
	"strconv"
	"strings"
)

func Recommend(target Movie, candidates []Movie, ctx UserContext) Recommendations {
	// Full O(n) scoring over ~5k titles is multi-second (unicode/root work).
	// Shortlist franchise + strong neighbors first, then score only those.
	candidates = shortlistCandidates(target, candidates, ctx)

	scored := make([]scoredMovie, len(candidates))
	for i, cm := range candidates {
		sim := ScoreSimilarContent(target, cm)
		scored[i] = scoredMovie{
			Movie:           cm,
			SeriesScore:     ScoreSeriesMatch(target, cm),
			SimilarScore:    sim,
			YouMayLikeScore: ScoreYouMayLike(target, cm, ctx),
		}
	}

	seen := make(map[string]bool)
	var sameSeries, similarContent, youMayLike []Movie

	sort.Slice(scored, func(i, j int) bool {
		a, b := scored[i], scored[j]
		if diff := a.SeriesScore - b.SeriesScore; diff > 0.01 {
			return true
		} else if diff < -0.01 {
			return false
		}
		return SeriesOrderLess(a.Movie, b.Movie)
	})
	for _, sm := range scored {
		if len(sameSeries) >= 30 {
			break
		}
		if sm.SeriesScore >= 10 && !seen[sm.Movie.Slug] {
			seen[sm.Movie.Slug] = true
			sameSeries = append(sameSeries, sm.Movie)
		}
	}

	sort.Slice(sameSeries, func(i, j int) bool {
		return SeriesOrderLess(sameSeries[i], sameSeries[j])
	})

	sort.Slice(scored, func(i, j int) bool {
		a, b := scored[i], scored[j]
		if diff := a.SimilarScore - b.SimilarScore; diff > 0.01 {
			return true
		} else if diff < -0.01 {
			return false
		}
		return SeriesOrderLess(a.Movie, b.Movie)
	})
	for _, sm := range scored {
		if len(similarContent) >= 12 {
			break
		}
		if sm.SimilarScore > 0 && !seen[sm.Movie.Slug] {
			seen[sm.Movie.Slug] = true
			similarContent = append(similarContent, sm.Movie)
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		a, b := scored[i], scored[j]
		if diff := a.YouMayLikeScore - b.YouMayLikeScore; diff > 0.01 {
			return true
		} else if diff < -0.01 {
			return false
		}
		return SeriesOrderLess(a.Movie, b.Movie)
	})
	for _, sm := range scored {
		if len(youMayLike) >= 8 {
			break
		}
		if seen[sm.Movie.Slug] || ctx.WatchedMovies[sm.Movie.Slug] {
			continue
		}
		if sm.YouMayLikeScore < 1.0 {
			continue
		}
		if IsSensitiveContent(sm.Movie) && !IsSensitiveContent(target) {
			continue
		}
		seen[sm.Movie.Slug] = true
		youMayLike = append(youMayLike, sm.Movie)
	}

	return Recommendations{
		SameSeries:     sameSeries,
		SimilarContent: similarContent,
		YouMayLike:     youMayLike,
	}
}

// shortlistCandidates keeps franchise hits + same-country strong genre neighbors.
// This cuts scoring from ~5k titles to a few hundred in typical catalogs.
func shortlistCandidates(target Movie, candidates []Movie, ctx UserContext) []Movie {
	if len(candidates) <= 400 {
		return candidates
	}

	tCountry := NormCountry(target.Country)
	tBaseEng := foldKey(seriesBase(target.OriginName))
	tBaseVi := foldKey(seriesBase(target.Name))
	tGenres := make(map[string]bool, len(target.Genres))
	for _, g := range target.Genres {
		tGenres[strings.ToLower(strings.TrimSpace(g))] = true
	}
	// slug family prefix (e.g. bat-kha-chien-bai, nguoi-nhen)
	tSlugFam := slugFamily(target.Slug)

	var franchise []Movie
	var neighbors []Movie
	seen := make(map[string]bool, 512)

	add := func(dst *[]Movie, m Movie) {
		if seen[m.Slug] {
			return
		}
		seen[m.Slug] = true
		*dst = append(*dst, m)
	}

	for _, cm := range candidates {
		if ctx.CoWatchedMovies[cm.Slug] {
			add(&franchise, cm)
			continue
		}

		blob := foldKey(cm.OriginName + " " + cm.Name + " " + cm.Slug)
		isFranchise := false
		if tBaseEng != "" && len(tBaseEng) >= 4 && strings.Contains(blob, tBaseEng) {
			isFranchise = true
		}
		if !isFranchise && tBaseVi != "" && len(tBaseVi) >= 4 && strings.Contains(blob, tBaseVi) {
			isFranchise = true
		}
		if !isFranchise && tSlugFam != "" && strings.HasPrefix(strings.ToLower(cm.Slug), tSlugFam) {
			isFranchise = true
		}
		// reverse: candidate's base appears in target (sequels with longer names)
		if !isFranchise {
			cBase := foldKey(seriesBase(cm.OriginName))
			if len(cBase) >= 5 {
				tBlob := foldKey(target.OriginName + " " + target.Name)
				if strings.Contains(tBlob, cBase) {
					isFranchise = true
				}
			}
		}
		if isFranchise {
			add(&franchise, cm)
			continue
		}

		if tCountry == "" || NormCountry(cm.Country) != tCountry {
			continue
		}
		overlap := 0
		for _, g := range cm.Genres {
			if tGenres[strings.ToLower(strings.TrimSpace(g))] {
				overlap++
			}
		}
		// Require real genre agreement so we don't score every same-country title.
		if overlap >= 2 {
			add(&neighbors, cm)
		}
	}

	// Cap same-country neighbors; franchise list is always kept in full.
	const maxNeighbors = 280
	if len(neighbors) > maxNeighbors {
		neighbors = neighbors[:maxNeighbors]
	}

	out := make([]Movie, 0, len(franchise)+len(neighbors))
	out = append(out, franchise...)
	out = append(out, neighbors...)
	if len(out) < 60 {
		// Degenerate metadata — fall back to full set rather than empty rails.
		return candidates
	}
	return out
}

func foldKey(s string) string {
	s = strings.ToLower(removeDiacritics(s))
	s = strings.NewReplacer("-", " ", "_", " ", ":", " ", ".", " ").Replace(s)
	return strings.Join(strings.Fields(s), " ")
}

func slugFamily(slug string) string {
	slug = strings.ToLower(slug)
	// keep prefix before -phan- / -season- / trailing number segments
	for _, sep := range []string{"-phan-", "-season-", "-ss-", "-tap-"} {
		if i := strings.Index(slug, sep); i > 2 {
			return slug[:i]
		}
	}
	parts := strings.Split(slug, "-")
	if len(parts) <= 2 {
		return slug
	}
	// drop trailing numeric tokens
	for len(parts) > 2 {
		last := parts[len(parts)-1]
		if _, err := strconv.Atoi(last); err == nil {
			parts = parts[:len(parts)-1]
			continue
		}
		break
	}
	if len(parts) > 4 {
		parts = parts[:4]
	}
	return strings.Join(parts, "-")
}
