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
		// Slug family is the safest franchise signal (bat-kha-chien-bai-*, nguoi-nhen-*).
		if tSlugFam != "" && strings.HasPrefix(strings.ToLower(cm.Slug), tSlugFam) {
			isFranchise = true
		}
		// Multi-token / distinctive bases may substring-match; single weak tokens
		// like "invincible" must use exact series-base equality only (not catalog-wide contains).
		if !isFranchise && tBaseEng != "" && len(tBaseEng) >= 4 {
			if isLooseFranchiseToken(tBaseEng) {
				if foldKey(seriesBase(cm.OriginName)) == tBaseEng || foldKey(seriesBase(cm.Name)) == tBaseEng {
					isFranchise = true
				}
			} else if strings.Contains(blob, tBaseEng) {
				isFranchise = true
			}
		}
		if !isFranchise && tBaseVi != "" && len(tBaseVi) >= 4 {
			if isLooseFranchiseToken(tBaseVi) {
				if foldKey(seriesBase(cm.Name)) == tBaseVi || foldKey(seriesBase(cm.OriginName)) == tBaseVi {
					isFranchise = true
				}
			} else if strings.Contains(blob, tBaseVi) {
				isFranchise = true
			}
		}
		// reverse: candidate's distinctive base appears in target
		if !isFranchise {
			cBase := foldKey(seriesBase(cm.OriginName))
			if len(cBase) >= 5 && !isLooseFranchiseToken(cBase) {
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

	out := make([]Movie, 0, len(franchise)+len(neighbors)+200)
	out = append(out, franchise...)
	out = append(out, neighbors...)

	// Never fall back to the full catalog (multi-second). If the shortlist is
	// thin, pad with same-country or shared-genre peers (capped).
	if len(out) < 80 {
		pad := padCandidates(target, candidates, seen, 200-len(out))
		out = append(out, pad...)
	}
	if len(out) == 0 {
		// Absolute last resort: first N candidates (still bounded).
		n := 150
		if len(candidates) < n {
			n = len(candidates)
		}
		return candidates[:n]
	}
	return out
}

func padCandidates(target Movie, candidates []Movie, seen map[string]bool, limit int) []Movie {
	if limit <= 0 {
		return nil
	}
	tCountry := NormCountry(target.Country)
	tGenres := make(map[string]bool, len(target.Genres))
	for _, g := range target.Genres {
		tGenres[strings.ToLower(strings.TrimSpace(g))] = true
	}
	var out []Movie
	for _, cm := range candidates {
		if seen[cm.Slug] {
			continue
		}
		overlap := 0
		for _, g := range cm.Genres {
			if tGenres[strings.ToLower(strings.TrimSpace(g))] {
				overlap++
			}
		}
		sameC := tCountry != "" && NormCountry(cm.Country) == tCountry
		if overlap >= 1 || sameC {
			seen[cm.Slug] = true
			out = append(out, cm)
			if len(out) >= limit {
				break
			}
		}
	}
	return out
}

func foldKey(s string) string {
	s = strings.ToLower(removeDiacritics(s))
	s = strings.NewReplacer("-", " ", "_", " ", ":", " ", ".", " ").Replace(s)
	return strings.Join(strings.Fields(s), " ")
}

// isLooseFranchiseToken is true for single-token bases that appear in many unrelated
// titles (substring shortlist would pull half the catalog).
func isLooseFranchiseToken(base string) bool {
	base = strings.TrimSpace(base)
	if base == "" {
		return true
	}
	if strings.Contains(base, " ") {
		// multi-token bases are usually distinctive enough for contains()
		// unless they're known weak phrases
		switch base {
		case "spider man", "iron man", "captain america", "wonder woman":
			return false // still ok via contains — many true peers
		}
		return false
	}
	// single token
	if isWeakContainmentToken(base) {
		return true
	}
	return len([]rune(base)) <= 6
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
