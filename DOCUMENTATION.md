# Technical documentation for movie recommendation system

This document provides a detailed look at the architecture, data structures and heuristic algorithms used in the movie recommendation system written in Go

## System overview

The system runs independently and does not depend on any external libraries
The goal is to compute and output a list of recommended movies split into three separate rows for the user based on a target movie

## Codebase structure

The matching logic is split into several modular files to ensure clean organization and maintainability
- `matching.go`: Contains the main entry point functions for computing matching scores (`ScoreSeriesMatch`, `ScoreSimilarContent`, `ScorePersonalised`, `getFranchiseMatchLevels`)
- `franchise.go`: Contains helper functions for parsing, extracting, and normalizing franchise roots (`extractRoots`, `franchiseRoots`, `vietnameseRoots`, `seriesBase`, `seriesSubtitle`, `slugBase`, `normRoot`, `isRootStopword`, `isTooGenericRoot`, `trimTrailingNumber`, `getPartNumber`)
- `similarity.go`: Contains functions to compute similarity profiles, animation format checks, and custom ordering (`rootOverlap`, `rootPrefixOverlap`, `rootCommonPrefix`, `meaningfulTokens`, `rootContains`, `SeriesOrderLess`, `sameAnimationProfile`, `animationFormat`, `animationProfile`, `isSeriesMovie`)
- `utils.go`: Contains general utility functions for name normalization and slice intersection count (`splitClean`, `normaliseName`, `intersectionCount`)

## Core data structures

### Movie
Describes the detailed information of a movie
- `ID`: Unique identifier of the movie
- `Slug`: Friendly URL slug of the movie
- `Name`: Vietnamese name or primary display name of the movie
- `OriginName`: Original name of the movie (usually in English)
- `PosterURL`: URL of the movie poster image
- `ThumbURL`: URL of the movie thumbnail image
- `Year`: Release year of the movie
- `Actors`: List of actors participating in the movie
- `Directors`: List of directors of the movie
- `Genres`: List of genres of the movie
- `Country`: Country of production
- `Content`: Content summary of the movie

### UserContext
Stores user behavior and preference information to personalize recommendations
- `GenreScores`: Score table of user's favorite genres
- `CoWatchedMovies`: List of movies commonly watched together
- `RecentGenres`: Statistics of recently watched movie genres
- `WatchedMovies`: List of movies the user has watched to filter out from recommendations appropriately

### Recommendations
The returned result contains three recommendation rows
- `SameSeries`: Movies in the same franchise or cinematic universe
- `SimilarContent`: Movies with similar content but not in the same franchise
- `YouMayLike`: Personalized recommendations specifically for the user

## Same series matching algorithm (Same Series)

The function `ScoreSeriesMatch` performs matching and scores to determine whether two movies belong to the same series
The minimum score to be considered as the same series is 10
The matching relies on the helper function `getFranchiseMatchLevels` to retrieve the English and Vietnamese match status along with the overall franchise level

### Franchise roots analysis (Franchise Roots)
- Extract the franchise root name from both the Vietnamese name (`Name`) and the original name (`OriginName`) using the helper function `extractRoots`
- Remove common stopwords (such as "phim", "movie", "season", "ss", "phan", "tap", "the", "and")
- Remove generic root names that are too broad (such as "love", "dark", "hero", "war") to avoid false positive matching
- Match the franchise root name between two movies (either exact match or partial/contained match)

### Animation and live-action format filtering
To avoid mixing up animated versions and live-action versions, the system classifies movies into three formats: `anime`, `cartoon` and `live-action`
- Animated movies from Japan or Korea are classified as `anime`
- Animated movies from other countries are classified as `cartoon`
- Other movies are classified as `live` (live-action)
- If two movies have different animation formats, 10 points will be deducted unless they share a common validated franchise root name

### Series order sorting
The function `SeriesOrderLess` defines the canonical sorting order of movies within the same series
- Prioritize sorting by release year in ascending order
- If release years are equal, TV shows or serials are sorted before movies (single films)
- If formats are equal, sort by part number (part/season) extracted from the name or slug
- Finally sort by alphabetical order of the movie name

## Content similarity algorithm (Similar Content)

The function `ScoreSimilarContent` calculates similarity based on content factors
- Retrieve matching information using the helper function `getFranchiseMatchLevels`
- Calculate the count of shared actors, shared directors and shared genres
- Add bonus points if two movies share a common franchise base name (for example, both belonging to the "Spider-Man" universe even if actors and directors are different)
- Apply a large penalty (-10 points) if two movies have different animation formats (for example, one animated anime and one live-action movie)

## Personalization algorithm (You May Like)

The function `ScorePersonalised` calculates scores based on user preferences and behavior
- Increase score based on the user's favorite genre scores for the movie genres
- Add bonus points if the movie is in the list of co-watched movies (`CoWatchedMovies`)
- Increase score based on the frequency of genres recently watched by the user
- Filter out movies the user has already watched

## Running tests

To ensure the algorithms work correctly and have no logic errors, run the following command
```bash
go test -v ./...
```
