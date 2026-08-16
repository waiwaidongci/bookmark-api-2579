package service

import "context"

func (s *Service) normalizeAndValidateBatchIDs(ctx context.Context, ids []int64) ([]int64, error) {
	if len(ids) == 0 {
		return nil, ErrNoFieldsToUpdate
	}

	seen := make(map[int64]struct{}, len(ids))
	normalized := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id < 1 {
			return nil, ErrInvalidID
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		normalized = append(normalized, id)
	}

	for _, id := range normalized {
		if _, err := s.repo.GetByID(ctx, id); err != nil {
			return nil, err
		}
	}
	return normalized, nil
}
