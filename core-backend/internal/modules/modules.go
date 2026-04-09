package modules

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/coregrpc"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam"
	"go.uber.org/fx"
)

var Modules = fx.Options(
	ai.Module,
	iam.Module,
	guide.Module,
	community.Module,
	coregrpc.Module,
)
