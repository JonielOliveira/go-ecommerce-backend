package domain

type Activatable struct {
	active bool
}

func NewActivatable() Activatable {
	return Activatable{active: true}
}

func NewActivatableFrom(active bool) Activatable {
	return Activatable{
		active: active,
	}
}

func (a *Activatable) IsActive() bool {
	return a.active
}

func (a *Activatable) Activate() {
	a.active = true
}

func (a *Activatable) Deactivate() {
	a.active = false
}
