package domain

import "time"

type SoftDelete struct {
	deletedAt *time.Time
}

func NewSoftDelete() SoftDelete {
	return SoftDelete{}
}

func NewSoftDeleteFrom(deletedAt *time.Time) SoftDelete {
	return SoftDelete{
		deletedAt: deletedAt,
	}
}

func (s *SoftDelete) DeletedAt() *time.Time {
	return s.deletedAt
}

func (s *SoftDelete) IsDeleted() bool {
	return s.deletedAt != nil
}

func (s *SoftDelete) Delete() {
	now := time.Now().UTC()
	s.deletedAt = &now
}

func (s *SoftDelete) Restore() {
	s.deletedAt = nil
}
