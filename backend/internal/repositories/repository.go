package repositories

import "gorm.io/gorm"

type Repository struct {
	EpicRepo         EpicRepository
	MessageRepo      MessageRepository
	PreviewTokenRepo PreviewTokenRepository
	ProjectRepo      ProjectRepository
	ProjectStageRepo StageRepository
	StoryRepo        StoryRepository
	UserRepo         UserRepository
}

func NewRepositories(db *gorm.DB) *Repository {
	return &Repository{
		EpicRepo:         NewEpicRepository(db),
		MessageRepo:      NewMessageRepository(db),
		PreviewTokenRepo: NewPreviewTokenRepository(db),
		ProjectRepo:      NewProjectRepository(db),
		ProjectStageRepo: NewStageRepository(db),
		StoryRepo:        NewStoryRepository(db),
		UserRepo:         NewUserRepository(db),
	}
}
