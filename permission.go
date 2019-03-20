package makeos

import "fmt"

// Permission represents a permission
type Permission struct {
	Actor string
	Level string
}

func (p Permission) String() string {
	return fmt.Sprintf("%s@%s", p.Actor, p.Level)
}
