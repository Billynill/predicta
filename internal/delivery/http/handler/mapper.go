package handler

import (
	"github.com/predicta/predicta/internal/delivery/http/dto"
	"github.com/predicta/predicta/internal/domain/entity"
)

func mapProjectStatus(s entity.ProjectStatus) dto.ProjectStatusResponse {
	return dto.ProjectStatusResponse{
		SprintName:    s.SprintName,
		CompletionPct: s.CompletionPct,
		DelayDays:     s.DelayDays,
		IsAtRisk:      s.IsAtRisk,
		RiskMessage:   s.RiskMessage,
		TrackName:     s.TrackName,
		DaysRemaining: s.DaysRemaining,
	}
}
