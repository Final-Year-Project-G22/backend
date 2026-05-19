package service

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
)

type seedTemplate struct {
	Slug           string
	Name           string
	DefaultTitle   string
	DefaultBody    string
	DefaultChannel *entity.Channel
}

func scheduledAlertTemplateSeedData() []seedTemplate {
	inApp := entity.ChannelInApp
	push := entity.ChannelPush

	return []seedTemplate{
		{Slug: "custom", Name: "Custom", DefaultTitle: "", DefaultBody: "", DefaultChannel: nil},
		{Slug: "tax_filing", Name: "Tax Filing Reminder", DefaultTitle: "Tax Filing Due", DefaultBody: "Your tax filing deadline is approaching. Make sure to submit your returns on time.", DefaultChannel: &inApp},
		{Slug: "license_renewal", Name: "License Renewal", DefaultTitle: "License Expiring", DefaultBody: "Your trade license renewal is due soon. Please prepare the required documents.", DefaultChannel: &inApp},
		{Slug: "registration_renewal", Name: "Registration Renewal", DefaultTitle: "Registration Expiring", DefaultBody: "Your business registration renewal is approaching. Check the expiry date and renew on time.", DefaultChannel: &inApp},
		{Slug: "meeting", Name: "Meeting Reminder", DefaultTitle: "Meeting Today", DefaultBody: "Don't forget your scheduled meeting.", DefaultChannel: &push},
		{Slug: "deadline", Name: "Custom Deadline", DefaultTitle: "Deadline Approaching", DefaultBody: "A deadline you set is coming up.", DefaultChannel: &inApp},
	}
}

type ScheduledAlertTemplateSeeder struct {
	tmplRepo notifrepo.ScheduledAlertTemplateRepository
	logger   core.Logger
}

func NewScheduledAlertTemplateSeeder(
	tmplRepo notifrepo.ScheduledAlertTemplateRepository,
	logger core.Logger,
) *ScheduledAlertTemplateSeeder {
	return &ScheduledAlertTemplateSeeder{
		tmplRepo: tmplRepo,
		logger:   logger,
	}
}

func (s *ScheduledAlertTemplateSeeder) Seed(ctx context.Context) error {
	_, err := s.tmplRepo.ListActive(ctx)
	if err != nil {
		s.logger.Info("Scheduled alert templates table not available, skipping seed",
			core.Error(err))
		return nil
	}

	templates := scheduledAlertTemplateSeedData()
	for _, t := range templates {
		existing, err := s.tmplRepo.GetBySlug(ctx, t.Slug)
		if err == nil && existing != nil {
			continue
		}

		tmpl := &entity.ScheduledAlertTemplate{
			Slug:           t.Slug,
			Name:           t.Name,
			DefaultTitle:   t.DefaultTitle,
			DefaultBody:    t.DefaultBody,
			DefaultChannel: t.DefaultChannel,
			IsActive:       true,
		}

		if err := s.tmplRepo.Create(ctx, tmpl); err != nil {
			return err
		}
		s.logger.Info("Scheduled alert template seeded", core.String("slug", tmpl.Slug))
	}
	return nil
}
