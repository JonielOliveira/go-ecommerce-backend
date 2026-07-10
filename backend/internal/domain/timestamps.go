package domain

import "time"

type Timestamps struct {
	createdAt time.Time
	updatedAt time.Time
}

func NewTimestamps() Timestamps {
	now := time.Now().UTC()
	return Timestamps{
		createdAt: now,
		updatedAt: now,
	}
}

func NewTimestampsFrom(createdAt, updatedAt time.Time) Timestamps {
	return Timestamps{
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (t *Timestamps) CreatedAt() time.Time {
	return t.createdAt
}

func (t *Timestamps) UpdatedAt() time.Time {
	return t.updatedAt
}

func (t *Timestamps) Touch() {
	t.updatedAt = time.Now().UTC()
}
