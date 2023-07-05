package authentication

import (
	"fmt"

	"github.com/casbin/casbin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Authorizer struct {
	enforcer *casbin.Enforcer
}

func NewAuthorizer(model, policy string) *Authorizer {
	enforcer := casbin.NewEnforcer(model, policy)
	return &Authorizer{enforcer: enforcer}
}

func (authorizer *Authorizer) Authorize(subject, object, action string) error {
	if !authorizer.enforcer.Enforce(subject, object, action) {
		message := fmt.Sprintf("%s not allowed to %s %s", subject, action, object)
		status := status.New(codes.PermissionDenied, message)
		return status.Err()
	}
	return nil
}
