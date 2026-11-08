package commands

import (
	"context"
	"testing"

	"github.com/Briyantama/HRIS-Stery/services/notification/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeConfigRepo struct {
	stored *domain.NotificationChannelConfig
}

func (f *fakeConfigRepo) Create(_ context.Context, c *domain.NotificationChannelConfig) error {
	f.stored = c
	return nil
}
func (f *fakeConfigRepo) GetByTenant(context.Context, domain.TenantID) (*domain.NotificationChannelConfig, error) {
	return f.stored, nil
}
func (f *fakeConfigRepo) Update(_ context.Context, c *domain.NotificationChannelConfig) error {
	f.stored = c
	return nil
}

func TestUpdateChannelConfig_KeepsOtherChannelsAndEmailConfig(t *testing.T) {
	repo := &fakeConfigRepo{}
	h := NewUpdateChannelConfigHandler(repo)
	tenant := uuid.NewString()
	email := &domain.EmailConfig{SMTPHost: "smtp.example.com", SMTPPort: 587, FromAddress: "noreply@example.com"}

	_, err := h.Handle(context.Background(), UpdateChannelConfigCommand{
		TenantID: tenant, Channels: []domain.ChannelType{domain.ChannelTypeEmail}, EmailConfig: email,
	})
	require.NoError(t, err)

	res, err := h.Handle(context.Background(), UpdateChannelConfigCommand{
		TenantID: tenant, Channels: []domain.ChannelType{domain.ChannelTypeInApp},
	})
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"IN_APP", "EMAIL"}, res.EnabledChannels)
	assert.True(t, repo.stored.IsChannelEnabled(domain.ChannelTypeEmail))
	assert.True(t, repo.stored.IsChannelEnabled(domain.ChannelTypeInApp))
	assert.Equal(t, email, repo.stored.GetEmailConfig())
}

func TestUpdateChannelConfig_EmailReplacesOnlyEmailConfig(t *testing.T) {
	repo := &fakeConfigRepo{}
	h := NewUpdateChannelConfigHandler(repo)
	tenant := uuid.NewString()
	old := &domain.EmailConfig{SMTPHost: "old.example.com", SMTPPort: 587, FromAddress: "a@example.com"}
	next := &domain.EmailConfig{SMTPHost: "new.example.com", SMTPPort: 465, FromAddress: "b@example.com"}

	for _, cmd := range []UpdateChannelConfigCommand{
		{TenantID: tenant, Channels: []domain.ChannelType{domain.ChannelTypeInApp}},
		{TenantID: tenant, Channels: []domain.ChannelType{domain.ChannelTypeEmail}, EmailConfig: old},
		{TenantID: tenant, Channels: []domain.ChannelType{domain.ChannelTypeEmail}, EmailConfig: next},
	} {
		_, err := h.Handle(context.Background(), cmd)
		require.NoError(t, err)
	}

	assert.True(t, repo.stored.IsChannelEnabled(domain.ChannelTypeInApp))
	assert.Equal(t, next, repo.stored.GetEmailConfig())
}
