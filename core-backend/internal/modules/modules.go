package modules

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam"
	"go.uber.org/fx"
)

var Modules = fx.Options(
	iam.Module,
)
