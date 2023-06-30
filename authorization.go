package objst

type AuthorizationHandler interface {
	IsAuthorized(owner string, id string) bool
}
