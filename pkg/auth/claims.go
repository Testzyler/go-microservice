package auth

import "github.com/google/uuid"

type Claims struct {
	Subject     uuid.UUID
	Roles       []string
	Permissions []string
}

func (c Claims) HasPermissions(required ...string) bool {
	return hasAll(c.Permissions, required)
}

func (c Claims) HasRoles(required ...string) bool {
	return hasAll(c.Roles, required)
}

func hasAll(have []string, required []string) bool {
	if len(required) == 0 {
		return true
	}
	if len(have) == 0 {
		return false
	}
	set := make(map[string]struct{}, len(have))
	for _, item := range have {
		set[item] = struct{}{}
	}
	for _, needed := range required {
		if _, ok := set[needed]; !ok {
			return false
		}
	}
	return true
}
