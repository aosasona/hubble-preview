package repository

import (
	"slices"
	"sort"
	"sync"

	"github.com/jackc/pgx/v5/pgtype"
	"go.trulyao.dev/hubble/web/internal/models"
)

const DefaultConcMinSize = 8

func ApplyMinThreshold(
	results []models.CollapsedSearchResult,
	threshold float64,
) []models.CollapsedSearchResult {
	// Sort the results again by their top percentile
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].RelevancePercent == 0 {
			return false
		}

		if results[j].RelevancePercent == 0 {
			return true
		}

		return results[i].RelevancePercent > results[j].RelevancePercent
	})

	// Remove entries that are below the threshold
	for i := len(results) - 1; i >= 0; i-- {
		if results[i].RelevancePercent < threshold {
			results = slices.Delete(results, i, i+1)
		}
	}

	return results
}

func collectIntoSortedSlice(
	resultsMap map[pgtype.UUID]models.CollapsedSearchResult,
) []models.CollapsedSearchResult {
	// Convert the map to a slice
	results := make([]models.CollapsedSearchResult, 0, len(resultsMap))
	for id := range resultsMap {
		result := resultsMap[id]
		results = append(results, result)
	}

	sortFn := func(i int) {
		result := &results[i]
		sort.SliceStable(result.Matches, func(a, b int) bool {
			return result.Matches[a].HybridScore > result.Matches[b].HybridScore
		})
	}

	// Concurrently sort the results by their top hybrid score (if the slice is at least 10 items)
	if len(results) >= DefaultConcMinSize {
		var wg sync.WaitGroup
		for i := range results {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				sortFn(i)
			}(i)
		}
		wg.Wait()
	} else {
		for i := range results {
			sortFn(i)
		}
	}

	// Re-sort the results again by their top hybrid score
	sort.SliceStable(results, func(i, j int) bool {
		if len(results[i].Matches) == 0 {
			return false
		}
		if len(results[j].Matches) == 0 {
			return true
		}

		return results[i].Matches[0].HybridScore > results[j].Matches[0].HybridScore
	})

	return results
}

func groupByEntry(
	chunkDedupMap map[int32]*models.SearchResult,
) (map[pgtype.UUID]models.CollapsedSearchResult, float64, float64) {
	var (
		results        = make(map[pgtype.UUID]models.CollapsedSearchResult)
		minHybridScore = 0.0
		maxHybridScore = 0.0
	)

	for _, result := range chunkDedupMap {
		entry, ok := results[result.ID]
		if !ok {
			results[result.ID] = models.CollapsedSearchResult{
				ID:            result.ID,
				Name:          result.Name,
				Type:          result.Type,
				Matches:       []models.MatchedChunk{},
				Status:        result.Status,
				Collection:    result.Collection,
				Workspace:     result.Workspace,
				CreatedAt:     result.CreatedAt,
				UpdatedAt:     result.UpdatedAt,
				ArchivedAt:    result.ArchivedAt,
				Metadata:      result.Metadata,
				FileID:        result.FileID,
				FilesizeBytes: result.FilesizeBytes,
			}
			entry = results[result.ID]
		}

		entry.Matches = append(entry.Matches, models.MatchedChunk{
			ID:            result.Chunk.ID,
			Text:          result.Preview,
			Rank:          result.Rank,
			Index:         result.Chunk.Index,
			TextScore:     result.TextScore,
			SemanticScore: result.SemanticScore,
			HybridScore:   result.HybridScore,
		})

		// Update min and max hybrid score
		if result.HybridScore > maxHybridScore {
			maxHybridScore = result.HybridScore
		}
		if result.HybridScore < minHybridScore || minHybridScore == 0 {
			minHybridScore = result.HybridScore
		}

		// Update the entry in the map
		results[result.ID] = entry
	}

	return results, minHybridScore, maxHybridScore
}

func calculateRelevancePercentile(
	results []models.CollapsedSearchResult,
) {
	if len(results) == 0 {
		return
	}

	// Calculate for the highest ranking entry
	maxHybridScore := results[0].Matches[0].HybridScore
	results[0].RelevancePercent = maxHybridScore / maxHybridScore * 100

	// Calculate for the other entries
	for i := 1; i < len(results); i++ {
		entry := &results[i]
		if len(entry.Matches) == 0 {
			continue
		}

		topScores := make([]float64, 0, len(entry.Matches))
		for matchIdx := range entry.Matches {
			match := entry.Matches[matchIdx]
			topScores = append(topScores, match.HybridScore)
		}
		avg := averageTopN(topScores, -1)

		entry.RelevancePercent = (avg / maxHybridScore) * 100
	}
}

func averageTopN(scores []float64, n int) float64 {
	var sum float64

	if len(scores) == 0 {
		return 0
	} else if n == -1 {
		n = len(scores)
	}

	if len(scores) < n {
		n = len(scores)
	}

	for i := range scores {
		if i >= n {
			break
		}
		sum += scores[i]
	}

	return sum / float64(n)
}
