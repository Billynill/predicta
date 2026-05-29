package port

import "context"

// UserLookup — аватар и профиль пользователя в task tracker (Jira).
type UserLookup interface {
	GetAvatarURL(ctx context.Context, accountID string) (string, error)
}
