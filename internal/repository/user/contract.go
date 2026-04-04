package user

import (
	"context"

	"github.com/internships-backend/test-backend-barashF/internal/repository/utils/transaction"
)

type manager interface {
	GetConn(ctx context.Context) (transaction.Connection, error)
}
