package station

import db "github.com/TeamOrchestrators/drishti-backend/db/db/generated"

// Service owns station business logic and sqlc access.
type Service struct{ queries *db.Queries }

func NewService(queries *db.Queries) *Service { return &Service{queries: queries} }
