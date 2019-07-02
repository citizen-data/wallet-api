package tenants

import (
	"context"
)

type TenantStore interface {
	GetTenantId(ctx context.Context, apikey string) (string, error)
}
