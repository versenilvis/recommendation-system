package recommender

import (
	"sort"
	"testing"
)

func TestGetCleanSeriesName(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"Người Nhện: Du Hành Vũ Trụ Nhện", "người nhện"},
		{"Người Nhện: Vũ Trụ Mới", "người nhện"},
		{"Người Nhện 2", "người nhện"},
		{"Người Nhện: Season 2", "người nhện"},
		{"Chú Thuật Hồi Chiến: Phần 2", "chú thuật hồi chiến"},
		{"Giờ Cao Điểm 3", "giờ cao điểm"},
		{"007: Skyfall", "007"},
	}

	for _, tc := range cases {
		got := seriesBase(tc.input)
		if got != tc.expected {
			t.Errorf("seriesBase(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

func TestIsSeries(t *testing.T) {
	cases := []struct {
		name     string
		slug     string
		genres   []string
		expected bool
	}{
		{"Người Nhện 2", "nguoi-nhen-2", []string{"phim-bo"}, true},
		{"Người Nhện 2", "nguoi-nhen-2", []string{"phim-le"}, false},
		{"Chú Thuật Hồi Chiến: Phần 2", "chu-thuat-hoi-chien-phan-2", []string{"anime"}, true},
		{"Chú Thuật Hồi Chiến Movie", "chu-thuat-hoi-chien-movie", []string{"anime"}, false},
	}

	for _, tc := range cases {
		m := Movie{Name: tc.name, Slug: tc.slug, Genres: tc.genres}
		got := isSeriesMovie(m)
		if got != tc.expected {
			t.Errorf("isSeriesMovie(%+v) = %t; want %t", m, got, tc.expected)
		}
	}
}

func TestGetPartNumber(t *testing.T) {
	cases := []struct {
		name     string
		slug     string
		expected int
	}{
		{"Người Nhện: Season 2", "nguoi-nhen-season-2", 2},
		{"Người Nhện: Phần 3", "nguoi-nhen-phan-3", 3},
		{"Người Nhện: ss4", "nguoi-nhen-ss4", 4},
		{"nguoi-nhen", "nguoi-nhen-phan-2", 2},
		{"Người Nhện Siêu Đẳng 2", "nguoi-nhen-sieu-dang-2", 2},
		{"nguoi-nhen-movie", "nguoi-nhen-movie", 0},
	}

	for _, tc := range cases {
		got := getPartNumber(tc.name, tc.slug)
		if got != tc.expected {
			t.Errorf("getPartNumber(%q, %q) = %d; want %d", tc.name, tc.slug, got, tc.expected)
		}
	}
}

func TestRecommendationSameSeriesScoringAndSorting(t *testing.T) {

	target := Movie{
		ID:         "1",
		Slug:       "nguoi-nhen-du-hanh-vu-tru-nhen",
		Name:       "Người Nhện: Du Hành Vũ Trụ Nhện",
		OriginName: "Spider-Man: Across the Spider-Verse",
		Year:       2023,
		Actors:     []string{"Shameik Moore", "Hailee Steinfeld"},
		Directors:  []string{"Joaquim Dos Santos"},
		Genres:     []string{"hoat-hinh", "hanh-dong"},
	}

	candidates := []Movie{
		{
			ID:         "2",
			Slug:       "nguoi-nhen-vu-tru-moi",
			Name:       "Người Nhện: Vũ Trụ Mới",
			OriginName: "Spider-Man: Into the Spider-Verse",
			Year:       2018,
			Actors:     []string{"Shameik Moore", "Jake Johnson"},
			Directors:  []string{"Bob Persichetti"},
			Genres:     []string{"hoat-hinh", "hanh-dong"},
		},
		{
			ID:         "3",
			Slug:       "nguoi-nhen-2",
			Name:       "Người Nhện 2",
			OriginName: "Spider-Man 2",
			Year:       2004,
			Actors:     []string{"Tobey Maguire", "Kirsten Dunst"},
			Directors:  []string{"Sam Raimi"},
			Genres:     []string{"hanh-dong", "phim-le"},
		},
		{
			ID:         "4",
			Slug:       "nguoi-nhen-3",
			Name:       "Người Nhện 3",
			OriginName: "Spider-Man 3",
			Year:       2007,
			Actors:     []string{"Tobey Maguire", "Kirsten Dunst"},
			Directors:  []string{"Sam Raimi"},
			Genres:     []string{"hanh-dong", "phim-le"},
		},
		{
			ID:         "5",
			Slug:       "nguoi-nhen-hanh-trinh-anh-hung",
			Name:       "Người Nhện: Hành Trình Anh Hùng",
			OriginName: "Spider-Man: Homecoming",
			Year:       2017,
			Actors:     []string{"Tom Holland", "Zendaya"},
			Directors:  []string{"Jon Watts"},
			Genres:     []string{"hanh-dong", "phim-le"},
		},
	}

	var scored []scoredMovie
	for _, cm := range candidates {
		seriesScore := ScoreSeriesMatch(target, cm)
		scored = append(scored, scoredMovie{
			Movie:        cm,
			SeriesScore:  seriesScore,
			SimilarScore: ScoreSimilarContent(target, cm),
		})
	}

	var spiderVerseScore, liveActionScore, liveActionSimilar float64
	for _, sm := range scored {
		if sm.Movie.ID == "2" {
			spiderVerseScore = sm.SeriesScore
		}
		if sm.Movie.ID == "3" {
			liveActionScore = sm.SeriesScore
			liveActionSimilar = sm.SimilarScore
		}
	}

	if spiderVerseScore < 10 {
		t.Errorf("expected animated Spider-Verse movie to qualify for same_series, got score %f", spiderVerseScore)
	}
	if liveActionScore != 0 {
		t.Errorf("expected live-action Spider-Man to be excluded from same_series for animated target, got score %f", liveActionScore)
	}
	if liveActionSimilar <= 0 {
		t.Errorf("expected live-action Spider-Man to remain eligible for similar_content, got score %f", liveActionSimilar)
	}

	sort.Slice(scored, func(i, j int) bool {
		a, b := scored[i], scored[j]
		if diff := a.SeriesScore - b.SeriesScore; diff > 0.01 {
			return true
		} else if diff < -0.01 {
			return false
		}
		return SeriesOrderLess(a.Movie, b.Movie)
	})

	if scored[0].Movie.ID != "2" {
		t.Errorf("expected first sorted movie to be ID 2, got ID %s", scored[0].Movie.ID)
	}
}

func TestRecommendationLiveActionSubSeriesUsesCastAndDirector(t *testing.T) {
	target := Movie{
		ID:         "1",
		Slug:       "nguoi-nhen-2",
		Name:       "Người Nhện 2",
		OriginName: "Spider-Man 2",
		Year:       2004,
		Actors:     []string{"Tobey Maguire", "Kirsten Dunst"},
		Directors:  []string{"Sam Raimi"},
		Genres:     []string{"hanh-dong", "phim-le"},
	}

	sameTrilogy := Movie{
		ID:         "2",
		Slug:       "nguoi-nhen-3",
		Name:       "Người Nhện 3",
		OriginName: "Spider-Man 3",
		Year:       2007,
		Actors:     []string{"Tobey Maguire", "Kirsten Dunst"},
		Directors:  []string{"Sam Raimi"},
		Genres:     []string{"hanh-dong", "phim-le"},
	}
	otherContinuity := Movie{
		ID:         "3",
		Slug:       "nguoi-nhen-hanh-trinh-anh-hung",
		Name:       "Người Nhện: Hành Trình Anh Hùng",
		OriginName: "Spider-Man: Homecoming",
		Year:       2017,
		Actors:     []string{"Tom Holland", "Zendaya"},
		Directors:  []string{"Jon Watts"},
		Genres:     []string{"hanh-dong", "phim-le"},
	}

	sameScore := ScoreSeriesMatch(target, sameTrilogy)
	otherScore := ScoreSeriesMatch(target, otherContinuity)

	if sameScore < 10 {
		t.Fatalf("expected same cast/director Spider-Man sequel to qualify for same_series, got score %f", sameScore)
	}
	if otherScore != 0 {
		t.Fatalf("expected different cast/director Spider-Man continuity to stay out of same_series, got score %f", otherScore)
	}
}

func TestRecommendationSpiderManOriginalTrilogyKeepsPartsTwoAndThree(t *testing.T) {
	target := Movie{
		ID:         "1",
		Slug:       "nguoi-nhen",
		Name:       "Người Nhện",
		OriginName: "Spider-Man",
		Year:       2002,
		Actors:     []string{"Tobey Maguire", "Kirsten Dunst"},
		Directors:  []string{"Sam Raimi"},
		Genres:     []string{"hanh-dong", "phim-le"},
	}
	part2 := Movie{
		ID:         "2",
		Slug:       "nguoi-nhen-2",
		Name:       "Người Nhện 2",
		OriginName: "Spider-Man 2",
		Year:       2004,
		Actors:     []string{"Tobey Maguire", "Kirsten Dunst"},
		Directors:  []string{"Sam Raimi"},
		Genres:     []string{"hanh-dong", "phim-le"},
	}
	part3 := Movie{
		ID:         "3",
		Slug:       "nguoi-nhen-3",
		Name:       "Người Nhện 3",
		OriginName: "Spider-Man 3",
		Year:       2007,
		Actors:     []string{"Tobey Maguire", "Kirsten Dunst"},
		Directors:  []string{"Sam Raimi"},
		Genres:     []string{"hanh-dong", "phim-le"},
	}
	reboot := Movie{
		ID:         "4",
		Slug:       "nguoi-nhen-tro-ve-nha",
		Name:       "Người Nhện: Trở Về Nhà",
		OriginName: "Spider-Man: Homecoming",
		Year:       2017,
		Actors:     []string{"Tom Holland", "Zendaya"},
		Directors:  []string{"Jon Watts"},
		Genres:     []string{"hanh-dong", "phim-le"},
	}

	part2Score := ScoreSeriesMatch(target, part2)
	part3Score := ScoreSeriesMatch(target, part3)
	rebootScore := ScoreSeriesMatch(target, reboot)
	if part2Score < 10 || part3Score < 10 {
		t.Fatalf("expected Spider-Man 2 and 3 to qualify, part2=%f part3=%f", part2Score, part3Score)
	}
	if rebootScore != 0 {
		t.Fatalf("expected reboot without actor/director overlap to stay out of same_series, got %f", rebootScore)
	}
}

func TestRecommendationAmazingSpiderManSequelUsesTrailingNumber(t *testing.T) {
	target := Movie{
		ID:         "1",
		Slug:       "nguoi-nhen-sieu-dang",
		Name:       "Người Nhện Siêu Đẳng",
		OriginName: "The Amazing Spider-Man",
		Year:       2012,
		Genres:     []string{"hanh-dong", "phim-le"},
	}
	sequel := Movie{
		ID:         "2",
		Slug:       "nguoi-nhen-sieu-dang-2",
		Name:       "Người Nhện Siêu Đẳng 2",
		OriginName: "The Amazing Spider-Man 2",
		Year:       2014,
		Genres:     []string{"hanh-dong", "phim-le"},
	}
	otherContinuity := Movie{
		ID:         "3",
		Slug:       "nguoi-nhen-tro-ve-nha",
		Name:       "Người Nhện: Trở Về Nhà",
		OriginName: "Spider-Man: Homecoming",
		Year:       2017,
		Genres:     []string{"hanh-dong", "phim-le"},
	}

	sequelScore := ScoreSeriesMatch(target, sequel)
	otherScore := ScoreSeriesMatch(target, otherContinuity)

	if sequelScore < 10 {
		t.Fatalf("expected Amazing Spider-Man 2 to qualify for same_series, got %f", sequelScore)
	}
	if otherScore != 0 {
		t.Fatalf("expected different Spider-Man continuity to stay out of same_series, got %f", otherScore)
	}
}

func TestRecommendationJujutsuKaisenKeepsSeasonAndMovieTogether(t *testing.T) {
	target := Movie{
		ID:         "1",
		Slug:       "chu-thuat-hoi-chien-phan-1",
		Name:       "Chú Thuật Hồi Chiến: Phần 1",
		OriginName: "Jujutsu Kaisen",
		Year:       2020,
		Genres:     []string{"anime", "hanh-dong"},
	}
	part2 := Movie{
		ID:         "2",
		Slug:       "chu-thuat-hoi-chien-phan-2",
		Name:       "Chú Thuật Hồi Chiến: Phần 2",
		OriginName: "Jujutsu Kaisen",
		Year:       2023,
		Genres:     []string{"anime", "hanh-dong"},
	}
	movie := Movie{
		ID:         "3",
		Slug:       "chu-thuat-hoi-chien-movie",
		Name:       "Chú Thuật Hồi Chiến Movie",
		OriginName: "Jujutsu Kaisen 0",
		Year:       2021,
		Genres:     []string{"anime", "hanh-dong"},
	}

	part2Score := ScoreSeriesMatch(target, part2)
	movieScore := ScoreSeriesMatch(target, movie)
	if part2Score < 10 || movieScore < 10 {
		t.Fatalf("expected Jujutsu Kaisen season/movie to qualify, part2=%f movie=%f", part2Score, movieScore)
	}
}

func TestOrderByReleaseThenPart(t *testing.T) {
	movies := []Movie{
		{Name: "Chú Thuật Hồi Chiến: Phần 2", Slug: "chu-thuat-hoi-chien-phan-2", Year: 2023, Genres: []string{"anime"}},
		{Name: "Chú Thuật Hồi Chiến: Phần 1", Slug: "chu-thuat-hoi-chien-phan-1", Year: 2020, Genres: []string{"anime"}},
		{Name: "Chú Thuật Hồi Chiến Movie", Slug: "chu-thuat-hoi-chien-movie", Year: 2021, Genres: []string{"anime"}},
	}

	sort.Slice(movies, func(i, j int) bool {
		return SeriesOrderLess(movies[i], movies[j])
	})

	got := []string{movies[0].Slug, movies[1].Slug, movies[2].Slug}
	want := []string{"chu-thuat-hoi-chien-phan-1", "chu-thuat-hoi-chien-movie", "chu-thuat-hoi-chien-phan-2"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v; want %v", got, want)
		}
	}
}

func TestNormaliseName(t *testing.T) {
	got := normaliseName("Natsuki Hanae")
	want := "hanae natsuki"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestIsTooGenericRootNewHeuristics(t *testing.T) {
	if isTooGenericRoot("larva") {
		t.Error("expected larva to be specific")
	}
	if isTooGenericRoot("bleach") {
		t.Error("expected bleach to be specific")
	}
	if !isTooGenericRoot("war") {
		t.Error("expected war to be generic")
	}
	if !isTooGenericRoot("love") {
		t.Error("expected love to be generic")
	}
}

func TestEnglishAndVietnameseMatchBoost(t *testing.T) {
	target := Movie{
		ID:         "1",
		Slug:       "thanh-guom-diet-quy-phan-1",
		Name:       "Thanh Gươm Diệt Quỷ (Phần 1)",
		OriginName: "Demon Slayer (Season 1)",
	}
	candidateFullMatch := Movie{
		ID:         "2",
		Slug:       "thanh-guom-diet-quy-phan-2",
		Name:       "Thanh Gươm Diệt Quỷ (Phần 2)",
		OriginName: "Demon Slayer (Season 2)",
	}
	candidateOnlyEngMatch := Movie{
		ID:         "3",
		Slug:       "thanh-guom-diet-quy-khac-nhau",
		Name:       "Diệt Quỷ Khác Nhau",
		OriginName: "Demon Slayer (Season 2)",
	}

	scoreFull := ScoreSeriesMatch(target, candidateFullMatch)
	scoreOnlyEng := ScoreSeriesMatch(target, candidateOnlyEngMatch)

	if scoreFull <= scoreOnlyEng {
		t.Errorf("expected scoreFull (%f) to be greater than scoreOnlyEng (%f)", scoreFull, scoreOnlyEng)
	}
}

func TestSingleTokenContainment(t *testing.T) {
	target := Movie{
		ID:         "1",
		Slug:       "au-trung-tinh-nghich-phan-3",
		Name:       "Ấu Trùng Tinh Nghịch (Phần 3)",
		OriginName: "Larva (Season 3)",
	}
	candidate := Movie{
		ID:         "2",
		Slug:       "dao-au-trung-phan-1",
		Name:       "Đảo Ấu Trùng (Phần 1)",
		OriginName: "Larva Island (Season 1)",
	}

	score := ScoreSeriesMatch(target, candidate)
	if score < 10 {
		t.Errorf("expected larva and larva island to match as same series, got score: %f", score)
	}
}

func TestCommonPrefixNonGenericSingleToken(t *testing.T) {
	target := Movie{
		ID:         "1",
		Slug:       "au-trung-tinh-nghich-mat-day-chuyen",
		Name:       "Ấu Trùng Tinh Nghịch: Mặt Dây Chuyền",
		OriginName: "Larva Pendant",
	}
	candidate := Movie{
		ID:         "2",
		Slug:       "dao-au-trung-phan-1",
		Name:       "Đảo Ấu Trùng (Phần 1)",
		OriginName: "Larva Island (Season 1)",
	}

	score := ScoreSeriesMatch(target, candidate)
	if score < 10 {
		t.Errorf("expected larva pendant and larva island to match as same series, got score: %f", score)
	}
}
