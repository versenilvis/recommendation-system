package recommender

import (
	"sort"
)

func Recommend(target Movie, candidates []Movie, ctx UserContext) Recommendations {
	scored := make([]scoredMovie, len(candidates))
	for i, cm := range candidates {
		scored[i] = scoredMovie{
			Movie:           cm,
			SeriesScore:     ScoreSeriesMatch(target, cm),
			SimilarScore:    ScoreSimilarContent(target, cm),
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
		// Require a positive similarity signal (filters pure noise + hard penalties).
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
		// Guests previously filled this rail with score-0 random rows (incl. 18+).
		// Require a meaningful positive score.
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
