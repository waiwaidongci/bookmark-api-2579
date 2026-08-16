package service

import (
	"sort"
	"strings"

	"bookmark-api/internal/model"
)

func normalizeTagStats(result model.TagStatsResult) model.TagStatsResult {
	if result.Tags == nil {
		result.Tags = make([]model.TagStat, 0)
	}
	for i := range result.Tags {
		result.Tags[i].Tag = strings.TrimSpace(result.Tags[i].Tag)
	}
	sort.Slice(result.Tags, func(i, j int) bool {
		return result.Tags[i].Tag < result.Tags[j].Tag
	})
	return result
}
