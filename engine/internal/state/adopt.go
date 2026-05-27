package state

import (
	"strings"
	"time"

	"github.com/hanchen1103/iteration_engine/domain"
)

func ApplyAdoptToVersion(version *domain.Version, actor string, now time.Time) {
	version.Status = domain.VersionStatusAdopted
	version.UpdatedBy = strings.TrimSpace(actor)
	version.UpdatedAt = now
}

func ApplyAdoptToRun(run *domain.Run, version *domain.Version, actor string, now time.Time) {
	run.Status = domain.RunStatusAdopted
	run.AdoptedVersionID = version.ID
	run.ActiveJobID = ""
	run.ActiveRoleKey = ""
	run.ActiveVersionID = ""
	run.UpdatedBy = strings.TrimSpace(actor)
	run.UpdatedAt = now
}
