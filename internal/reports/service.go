package reports

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) TeamStats(ctx context.Context) (*TeamStatsResponse, error) {
	items, err := s.repo.TeamStats(ctx)

	if err != nil {
		return nil, err
	}

	return &TeamStatsResponse{
		Items: items,
	}, nil
}

func (s *Service) TopCreators(ctx context.Context) (*TopCreatorsResponse, error) {
	items, err := s.repo.TopCreators(ctx)

	if err != nil {
		return nil, err
	}

	return &TopCreatorsResponse{
		Items: items,
	}, nil
}

func (s *Service) InvalidAssignees(ctx context.Context) (*InvalidAssigneesResponse, error) {
	items, err := s.repo.InvalidAssignees(ctx)

	if err != nil {
		return nil, err
	}

	return &InvalidAssigneesResponse{
		Items: items,
	}, nil
}