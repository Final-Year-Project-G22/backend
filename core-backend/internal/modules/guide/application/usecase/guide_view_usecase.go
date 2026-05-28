package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/entity"
	guideerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/usecase"
	iamrepository "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	notifentity "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notifevent "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/event"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/constants"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/notificationevent"
	sharedRepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type guideViewUsecase struct {
	guideRepo       repository.GuideRepository
	stepRepo        repository.StepRepository
	progressRepo    repository.ProgressRepository
	profileRepo     iamrepository.BusinessProfileRepository
	transactor      sharedRepo.Transactor
	notifOutboxRepo notifrepo.NotificationOutboxRepository
	logger          core.Logger
}

func NewGuideViewUsecase(
	guideRepo repository.GuideRepository,
	stepRepo repository.StepRepository,
	progressRepo repository.ProgressRepository,
	profileRepo iamrepository.BusinessProfileRepository,
	transactor sharedRepo.Transactor,
	notifOutboxRepo notifrepo.NotificationOutboxRepository,
	logger core.Logger,
) usecase.GuideViewUseCase {
	return &guideViewUsecase{
		guideRepo:       guideRepo,
		stepRepo:        stepRepo,
		progressRepo:    progressRepo,
		profileRepo:     profileRepo,
		transactor:      transactor,
		notifOutboxRepo: notifOutboxRepo,
		logger:          logger,
	}
}

func (s *guideViewUsecase) ListGuides(ctx context.Context, accountID, userID uuid.UUID, q query.QueryOptions, locale constants.Locale, sectorIDs, tagIDs []uuid.UUID) ([]*usecase.GuideCard, error) {
	// If explicit taxonomy filters are provided, use them directly
	if len(sectorIDs) == 0 && len(tagIDs) == 0 {
		// Fetch user's business profile for taxonomy filtering
		profile, err := s.profileRepo.GetByAccountID(ctx, accountID)
		if err != nil {
			s.logger.Error("Failed to get business profile for guide filtering", core.Error(err))
			// If profile not found, return all guides (no restriction)
		}

		if profile != nil {
			// Build sector filter: include profile's sector and its ancestors
			if profile.SectorID != nil {
				sectorIDs = append(sectorIDs, *profile.SectorID)
			}
			// TODO: include ancestor sectors from sector repository when subtree lookup is available
			for _, tag := range profile.Tags {
				tagIDs = append(tagIDs, tag.ID)
			}
		}
	}

	guides, err := s.guideRepo.ListByTaxonomy(ctx, sectorIDs, tagIDs, q, locale)
	if err != nil {
		return nil, err
	}

	cards := make([]*usecase.GuideCard, len(guides))
	for i, g := range guides {
		cards[i] = s.toGuideCard(g)
	}
	return cards, nil
}

func (s *guideViewUsecase) ListAllGuides(ctx context.Context, q query.QueryOptions, locale constants.Locale) ([]*usecase.GuideCard, error) {
	guides, err := s.guideRepo.ListByTaxonomy(ctx, nil, nil, q, locale)
	if err != nil {
		return nil, err
	}

	cards := make([]*usecase.GuideCard, len(guides))
	for i, g := range guides {
		cards[i] = s.toGuideCard(g)
	}
	return cards, nil
}

func (s *guideViewUsecase) SearchGuides(ctx context.Context, accountID, userID uuid.UUID, keyword string, q query.QueryOptions, locale constants.Locale) ([]*usecase.GuideCard, error) {
	guides, err := s.guideRepo.Search(ctx, keyword, q, locale)
	if err != nil {
		return nil, err
	}
	cards := make([]*usecase.GuideCard, len(guides))
	for i, g := range guides {
		cards[i] = s.toGuideCard(g)
	}
	return cards, nil
}

func (s *guideViewUsecase) GetInProgressGuides(ctx context.Context, accountID, userID uuid.UUID, locale constants.Locale) ([]*usecase.GuideWithProgress, error) {
	guides, completedCounts, totalCounts, err := s.progressRepo.ListGuidesInProgress(ctx, accountID, userID, locale)
	if err != nil {
		return nil, err
	}
	result := make([]*usecase.GuideWithProgress, 0, len(guides))
	for i, g := range guides {
		name := g.Slug
		if len(g.Translations) > 0 {
			name = g.Translations[0].Name
		}
		result = append(result, &usecase.GuideWithProgress{
			ID:             g.ID,
			Slug:           g.Slug,
			Name:           name,
			Icon:           g.Icon,
			CompletedSteps: completedCounts[i],
			TotalSteps:     totalCounts[i],
		})
	}
	return result, nil
}

func (s *guideViewUsecase) GetCompletionStats(ctx context.Context, accountID, userID uuid.UUID) (*usecase.CompletionStats, error) {
	completedGuides, inProgressGuides, totalStepsCompleted, totalStepsAll, err := s.progressRepo.GetProgressStats(ctx, accountID, userID)
	if err != nil {
		return nil, err
	}
	return &usecase.CompletionStats{
		CompletedGuides:     completedGuides,
		InProgressGuides:    inProgressGuides,
		TotalStepsCompleted: totalStepsCompleted,
		TotalStepsAll:       totalStepsAll,
		Period:              "monthly",
	}, nil
}

func (s *guideViewUsecase) GetRecentlyViewed(ctx context.Context, accountID, userID uuid.UUID, q query.QueryOptions, locale constants.Locale) ([]*usecase.GuideCard, error) {
	if q.PageSize < 1 {
		q.PageSize = 5
	}
	if q.PageSize > 5 {
		q.PageSize = 5
	}
	guides, err := s.progressRepo.ListRecentlyViewedGuides(ctx, accountID, userID, q, locale)
	if err != nil {
		return nil, err
	}
	cards := make([]*usecase.GuideCard, len(guides))
	for i, g := range guides {
		cards[i] = s.toGuideCard(g)
	}
	return cards, nil
}

func (s *guideViewUsecase) GetPersonalizedGuide(ctx context.Context, accountID, userID uuid.UUID, guideSlug string, locale constants.Locale) (*usecase.PersonalizedGuide, error) {
	guide, err := s.guideRepo.GetBySlugGlobal(ctx, guideSlug, locale)
	if err != nil {
		return nil, err
	}

	steps, err := s.stepRepo.ListByGuide(ctx, guide.ID, query.QueryOptions{PageSize: 200}, locale)
	if err != nil {
		return nil, err
	}

	progressList, err := s.progressRepo.ListProgressByGuide(ctx, accountID, userID, guide.ID, query.QueryOptions{PageSize: 200})
	if err != nil {
		return nil, err
	}

	progressMap := make(map[uuid.UUID]*entity.UserGuideProgress, len(progressList))
	for _, p := range progressList {
		progressMap[p.StepID] = p
	}

	personalizedSteps := make([]*usecase.PersonalizedStep, len(steps))
	var completed, skipped, inProgress int
	var estimatedTotal int
	hasEstimate := false
	for i, step := range steps {
		p, exists := progressMap[step.ID]
		status := entity.ProgressStatusLocked
		if exists {
			status = p.Status
		}
		switch status {
		case entity.ProgressStatusCompleted:
			completed++
		case entity.ProgressStatusSkipped:
			skipped++
		case entity.ProgressStatusInProgress:
			inProgress++
		}
		if step.EstimatedTime != nil {
			estimatedTotal += *step.EstimatedTime
			hasEstimate = true
		}
		personalizedSteps[i] = s.toPersonalizedStep(step, status)
	}

	var estimatedTotalPtr *int
	if hasEstimate {
		estimatedTotalPtr = &estimatedTotal
	}
	computedHash := journeyHashForSteps(steps)

	existingJourney, err := s.progressRepo.GetJourney(ctx, accountID, userID, guide.ID)
	journeyNeedsUpdate := err != nil || existingJourney == nil || existingJourney.JourneyHash == nil || *existingJourney.JourneyHash != *computedHash

	if journeyNeedsUpdate {
		journey := &entity.UserGuideJourney{
			AccountID:          accountID,
			UserID:             userID,
			GuideID:            guide.ID,
			JourneyHash:        computedHash,
			StepSequence:       stepsToSequenceJSONMap(steps),
			TotalSteps:         len(steps),
			CompletedSteps:     completed,
			EstimatedTotalTime: estimatedTotalPtr,
			GeneratedAt:        time.Now().UTC(),
		}
		if err := s.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
			if err := s.progressRepo.UpsertRecentView(txCtx, accountID, userID, guide.ID); err != nil {
				return err
			}
			return s.progressRepo.UpsertJourney(txCtx, journey)
		}); err != nil {
			s.logger.Error("Failed to create journey", core.Error(err))
		}
	} else {
		if err := s.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
			return s.progressRepo.UpsertRecentView(txCtx, accountID, userID, guide.ID)
		}); err != nil {
			s.logger.Error("Failed to update recent view", core.Error(err))
		}
	}

	return &usecase.PersonalizedGuide{
		ID:          guide.ID,
		Slug:        guide.Slug,
		Name:        s.resolveGuideName(guide, locale),
		Description: s.resolveGuideDescription(guide, locale),
		Steps:       personalizedSteps,
		Progress: &usecase.GuideProgressSummary{
			TotalSteps:      len(steps),
			CompletedSteps:  completed,
			SkippedSteps:    skipped,
			InProgressSteps: inProgress,
		},
	}, nil
}

func (s *guideViewUsecase) GetCurrentStep(ctx context.Context, accountID, userID uuid.UUID, guideSlug string, locale constants.Locale) (*usecase.GetCurrentStepResult, error) {
	guide, err := s.guideRepo.GetBySlugGlobal(ctx, guideSlug, locale)
	if err != nil {
		return nil, err
	}

	steps, err := s.stepRepo.ListByGuide(ctx, guide.ID, query.QueryOptions{PageSize: 200}, locale)
	if err != nil {
		return nil, err
	}

	progressList, err := s.progressRepo.ListProgressByGuide(ctx, accountID, userID, guide.ID, query.QueryOptions{PageSize: 200})
	if err != nil {
		return nil, err
	}

	progressMap := make(map[uuid.UUID]entity.ProgressStatus, len(progressList))
	for _, p := range progressList {
		progressMap[p.StepID] = p.Status
	}

	for _, step := range steps {
		status, exists := progressMap[step.ID]
		if !exists || status == entity.ProgressStatusLocked {
			return &usecase.GetCurrentStepResult{
				ID:            step.ID,
				Slug:          step.Slug,
				Title:         s.resolveStepTitle(step, locale),
				Description:   s.resolveStepDescription(step, locale),
				StepType:      step.StepType,
				SortOrder:     step.SortOrder,
				IsOptional:    step.IsOptional,
				EstimatedTime: step.EstimatedTime,
			}, nil
		}
	}

	return nil, nil
}

func (s *guideViewUsecase) StartStep(ctx context.Context, accountID, userID, stepID uuid.UUID) error {
	progress, err := s.progressRepo.GetProgress(ctx, accountID, userID, stepID)
	if err != nil && err != guideerror.ErrProgressNotFound {
		return err
	}

	if progress != nil {
		if progress.Status == entity.ProgressStatusInProgress {
			return nil
		}
		if progress.Status != entity.ProgressStatusLocked {
			return errors.ConflictError("guide.errors.invalidStatusTransition")
		}
	}

	deps, err := s.stepRepo.GetDependencies(ctx, stepID)
	if err != nil {
		return err
	}
	for _, dep := range deps {
		depProgress, err := s.progressRepo.GetProgress(ctx, accountID, userID, dep.RequiredStepID)
		if err != nil || depProgress == nil || depProgress.Status != entity.ProgressStatusCompleted {
			return errors.BadRequestError("guide.errors.dependenciesNotMet")
		}
	}

	now := time.Now().UTC()
	if progress == nil {
		progress = &entity.UserGuideProgress{
			AccountID: accountID,
			UserID:    userID,
			StepID:    stepID,
			Status:    entity.ProgressStatusInProgress,
			StartedAt: &now,
		}
	} else {
		progress.Status = entity.ProgressStatusInProgress
		progress.StartedAt = &now
	}

	return s.progressRepo.UpsertProgress(ctx, progress)
}

func (s *guideViewUsecase) CompleteStep(ctx context.Context, accountID, userID, stepID uuid.UUID, input usecase.CompleteStepInput) error {
	progress, err := s.progressRepo.GetProgress(ctx, accountID, userID, stepID)
	if err != nil {
		return err
	}
	if progress.Status != entity.ProgressStatusInProgress {
		return errors.ConflictError("guide.errors.invalidStatusTransition")
	}

	now := time.Now().UTC()
	progress.Status = entity.ProgressStatusCompleted
	progress.CompletedAt = &now
	if input.TimeSpentSeconds != nil {
		progress.TimeSpent = input.TimeSpentSeconds
	}
	if input.Notes != nil {
		progress.Notes = input.Notes
	}
	if len(input.UploadedDocuments) > 0 {
		progress.UploadedDocuments = documentsToJSONMap(input.UploadedDocuments)
	}

	if err := s.progressRepo.UpsertProgress(ctx, progress); err != nil {
		return err
	}

	return s.publishStepEvent(ctx, stepID, accountID, notifevent.GuideComplianceStepCompleted)
}

func (s *guideViewUsecase) MarkStepIncomplete(ctx context.Context, accountID, userID, stepID uuid.UUID) error {
	progress, err := s.progressRepo.GetProgress(ctx, accountID, userID, stepID)
	if err != nil {
		return err
	}
	if progress.Status != entity.ProgressStatusCompleted {
		return errors.ConflictError("guide.errors.invalidStatusTransition")
	}

	progress.Status = entity.ProgressStatusInProgress
	progress.CompletedAt = nil

	if err := s.progressRepo.UpsertProgress(ctx, progress); err != nil {
		return err
	}

	return s.publishStepEvent(ctx, stepID, accountID, notifevent.GuideComplianceStepRolledBack)
}

func (s *guideViewUsecase) publishStepEvent(ctx context.Context, stepID, accountID uuid.UUID, eventType string) error {
	step, err := s.stepRepo.GetByID(ctx, stepID)
	if err != nil {
		s.logger.Warn("Failed to load step for compliance event", core.String("stepID", stepID.String()))
		return nil
	}

	if step.ComplianceType == nil || *step.ComplianceType == "" {
		return nil
	}

	variables := map[string]string{
		"compliance_type": *step.ComplianceType,
	}

	env := notificationevent.Envelope{
		SchemaVersion:    notificationevent.SchemaVersionV1,
		EventType:        eventType,
		OccurredAt:       time.Now().UTC(),
		SourceModule:     "guide",
		AccountID:        accountID,
		NotificationType: "guide_step_completed",
		ChannelPolicy:    notificationevent.ChannelPolicySingle,
		Channel:          strPtr("in_app"),
		Variables:        variables,
		Metadata: notificationevent.Metadata{
			IdempotencyKey: eventType + ":" + stepID.String() + ":" + accountID.String() + ":" + uuid.New().String(),
		},
	}

	payload := make(map[string]interface{})
	data, _ := json.Marshal(env)
	_ = json.Unmarshal(data, &payload)

	outbox := &notifentity.NotificationOutbox{
		EventType:      env.EventType,
		SchemaVersion:  env.SchemaVersion,
		SourceModule:   env.SourceModule,
		AccountID:      env.AccountID,
		IdempotencyKey: env.Metadata.IdempotencyKey,
		Payload:        payload,
		Status:         notifentity.NotificationOutboxStatusPending,
	}
	return s.notifOutboxRepo.Create(ctx, outbox)
}

func strPtr(s string) *string {
	return &s
}

func (s *guideViewUsecase) SkipOptionalStep(ctx context.Context, accountID, userID, stepID uuid.UUID, reason *string) error {
	step, err := s.stepRepo.GetByID(ctx, stepID)
	if err != nil {
		return err
	}
	if !step.IsOptional {
		return errors.BadRequestError("guide.errors.stepNotOptional")
	}

	progress, err := s.progressRepo.GetProgress(ctx, accountID, userID, stepID)
	if err != nil {
		return err
	}
	if progress.Status != entity.ProgressStatusInProgress {
		return errors.ConflictError("guide.errors.invalidStatusTransition")
	}

	now := time.Now().UTC()
	progress.Status = entity.ProgressStatusSkipped
	progress.CompletedAt = &now
	if reason != nil {
		progress.Notes = reason
	}

	return s.progressRepo.UpsertProgress(ctx, progress)
}

func (s *guideViewUsecase) UpdateStepProgress(ctx context.Context, accountID, userID, stepID uuid.UUID, input usecase.UpdateProgressInput) error {
	progress, err := s.progressRepo.GetProgress(ctx, accountID, userID, stepID)
	if err != nil && err != guideerror.ErrProgressNotFound {
		return err
	}

	if progress == nil {
		now := time.Now().UTC()
		progress = &entity.UserGuideProgress{
			AccountID: accountID,
			UserID:    userID,
			StepID:    stepID,
			Status:    entity.ProgressStatusInProgress,
			StartedAt: &now,
		}
	}

	if input.TimeSpentSeconds != nil {
		progress.TimeSpent = input.TimeSpentSeconds
	}
	if input.Notes != nil {
		progress.Notes = input.Notes
	}
	if len(input.UploadedDocuments) > 0 {
		progress.UploadedDocuments = documentsToJSONMap(input.UploadedDocuments)
	}
	now := time.Now().UTC()
	progress.LastAccessedAt = &now

	return s.progressRepo.UpsertProgress(ctx, progress)
}

func (s *guideViewUsecase) AddBookmark(ctx context.Context, accountID, userID, stepID uuid.UUID, note *string) error {
	bookmark := &entity.UserGuideBookmark{
		AccountID: accountID,
		UserID:    userID,
		StepID:    stepID,
		Note:      note,
	}
	return s.progressRepo.UpsertBookmark(ctx, bookmark)
}

func (s *guideViewUsecase) UpdateBookmarkNote(ctx context.Context, accountID, userID, stepID uuid.UUID, note *string) error {
	bookmark, err := s.progressRepo.GetBookmark(ctx, accountID, userID, stepID)
	if err != nil {
		return err
	}
	bookmark.Note = note
	return s.progressRepo.UpsertBookmark(ctx, bookmark)
}

func (s *guideViewUsecase) RemoveBookmark(ctx context.Context, accountID, userID, stepID uuid.UUID) error {
	return s.progressRepo.RemoveBookmark(ctx, accountID, userID, stepID)
}

func (s *guideViewUsecase) ListBookmarks(ctx context.Context, accountID, userID uuid.UUID, q query.QueryOptions, locale constants.Locale) ([]*usecase.BookmarkWithStep, error) {
	bookmarks, err := s.progressRepo.ListBookmarks(ctx, accountID, userID, q)
	if err != nil {
		return nil, err
	}
	result := make([]*usecase.BookmarkWithStep, len(bookmarks))
	for i, b := range bookmarks {
		stepTitle := ""
		guideName := ""
		if b.Step.ID != uuid.Nil {
			stepTitle = s.resolveStepTitle(&b.Step, locale)
			guideName = s.resolveGuideName(&b.Step.Guide, locale)
		}
		result[i] = &usecase.BookmarkWithStep{
			ID:        b.ID,
			StepID:    b.StepID,
			Note:      b.Note,
			StepTitle: stepTitle,
			GuideName: guideName,
			CreatedAt: b.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}
	return result, nil
}

func (s *guideViewUsecase) toGuideCard(g *entity.Guide) *usecase.GuideCard {
	name := g.Slug
	var desc *string
	if len(g.Translations) > 0 {
		t := g.Translations[0]
		name = t.Name
		desc = t.Description
	}
	return &usecase.GuideCard{
		ID:          g.ID,
		Slug:        g.Slug,
		Name:        name,
		Description: desc,
		Icon:        g.Icon,
		SectorIDs:   g.SectorIDs,
		TagIDs:      g.TagIDs,
	}
}

func (s *guideViewUsecase) toPersonalizedStep(step *entity.GuideStep, status entity.ProgressStatus) *usecase.PersonalizedStep {
	title := step.Slug
	var desc *string
	var detailedContent map[string]interface{}
	if len(step.Translations) > 0 {
		t := step.Translations[0]
		title = t.Title
		desc = t.Description
		if t.DetailedContent != nil {
			detailedContent = map[string]interface{}(t.DetailedContent)
		}
	}
	return &usecase.PersonalizedStep{
		ID:              step.ID,
		Slug:            step.Slug,
		Title:           title,
		Description:     desc,
		StepType:        step.StepType,
		SortOrder:       step.SortOrder,
		IsOptional:      step.IsOptional,
		Status:          status,
		EstimatedTime:   step.EstimatedTime,
		DetailedContent: detailedContent,
	}
}

func (s *guideViewUsecase) resolveGuideName(g *entity.Guide, locale constants.Locale) string {
	for _, t := range g.Translations {
		if t.Language == string(locale) {
			return t.Name
		}
	}
	if len(g.Translations) > 0 {
		return g.Translations[0].Name
	}
	return g.Slug
}

func (s *guideViewUsecase) resolveGuideDescription(g *entity.Guide, locale constants.Locale) *string {
	for _, t := range g.Translations {
		if t.Language == string(locale) {
			return t.Description
		}
	}
	if len(g.Translations) > 0 {
		return g.Translations[0].Description
	}
	return nil
}

func (s *guideViewUsecase) resolveStepTitle(step *entity.GuideStep, locale constants.Locale) string {
	for _, t := range step.Translations {
		if t.Language == string(locale) {
			return t.Title
		}
	}
	if len(step.Translations) > 0 {
		return step.Translations[0].Title
	}
	return step.Slug
}

func (s *guideViewUsecase) resolveStepDescription(step *entity.GuideStep, locale constants.Locale) *string {
	for _, t := range step.Translations {
		if t.Language == string(locale) {
			return t.Description
		}
	}
	if len(step.Translations) > 0 {
		return step.Translations[0].Description
	}
	return nil
}

func documentsToJSONMap(documents []string) datatypes.JSON {
	if len(documents) == 0 {
		return datatypes.JSON("{}")
	}
	result := datatypes.JSONMap{}
	for i, doc := range documents {
		result[strconv.Itoa(i)] = doc
	}
	b, _ := json.Marshal(result)
	return datatypes.JSON(b)
}

func stepsToSequenceJSONMap(steps []*entity.GuideStep) datatypes.JSON {
	if len(steps) == 0 {
		return datatypes.JSON("{}")
	}
	result := datatypes.JSONMap{}
	for i, step := range steps {
		result[strconv.Itoa(i)] = step.ID.String()
	}
	b, _ := json.Marshal(result)
	return datatypes.JSON(b)
}

func journeyHashForSteps(steps []*entity.GuideStep) *string {
	if len(steps) == 0 {
		return nil
	}
	parts := make([]string, 0, len(steps))
	for _, step := range steps {
		parts = append(parts, step.ID.String()+":"+strconv.Itoa(step.Version))
	}
	joined := strings.Join(parts, "|")
	hash := sha256.Sum256([]byte(joined))
	encoded := hex.EncodeToString(hash[:])
	return &encoded
}
