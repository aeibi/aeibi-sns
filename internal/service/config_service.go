package service

import (
	"aeibi/api"
	"aeibi/internal/config"
	"context"
)

type ConfigService struct {
	maxUploadSizeKB int32
	siteName        string
	siteIconURL     string
	siteLogoURL     string
}

func NewConfigService(cfg *config.Config) *ConfigService {
	return &ConfigService{
		maxUploadSizeKB: int32(cfg.OSS.MaxUploadSizeKB),
		siteName:        cfg.Frontend.SiteName,
		siteIconURL:     cfg.Frontend.SiteIconURL,
		siteLogoURL:     cfg.Frontend.SiteLogoURL,
	}
}

func (s *ConfigService) GetFrontendConfig(_ context.Context, _ *api.GetFrontendConfigRequest) (*api.GetFrontendConfigResponse, error) {
	return &api.GetFrontendConfigResponse{
		Config: &api.FrontendConfig{
			SiteName:    s.siteName,
			SiteIconUrl: s.siteIconURL,
			SiteLogoUrl: s.siteLogoURL,
		},
	}, nil
}

func (s *ConfigService) GetUploadConfig(_ context.Context, _ *api.GetUploadConfigRequest) (*api.GetUploadConfigResponse, error) {
	return &api.GetUploadConfigResponse{
		Config: &api.UploadConfig{
			MaxUploadSizeKb: s.maxUploadSizeKB,
		},
	}, nil
}
