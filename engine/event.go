package engine

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/hanchen1103/iteration_engine/domain"
)

func (s *Service) recordEvent(ctx context.Context, eventType string, runID string, versionID string, actor string, message string, payload json.RawMessage) {
	if s == nil || s.store == nil || strings.TrimSpace(runID) == "" {
		return
	}
	event := &domain.Event{
		RunID:     strings.TrimSpace(runID),
		VersionID: strings.TrimSpace(versionID),
		Type:      eventType,
		Actor:     strings.TrimSpace(actor),
		Message:   strings.TrimSpace(message),
		Payload:   domain.CloneRawMessage(payload),
		CreatedAt: s.now(),
	}
	_ = s.store.RecordEvent(ctx, event)
}
