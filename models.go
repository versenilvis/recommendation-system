package recommender

type Movie struct {
	ID         string
	Slug       string
	Name       string
	OriginName string
	PosterURL  string
	ThumbURL   string
	Year       int
	Actors     []string
	Directors  []string
	Genres     []string
	Country    string
	Content    string
}

type UserContext struct {
	GenreScores     map[string]float64
	CoWatchedMovies map[string]bool
	RecentGenres    map[string]int
	WatchedMovies   map[string]bool
}

type Recommendations struct {
	SameSeries     []Movie
	SimilarContent []Movie
	YouMayLike     []Movie
}

type scoredMovie struct {
	Movie           Movie
	SeriesScore     float64
	SimilarScore    float64
	YouMayLikeScore float64
}
