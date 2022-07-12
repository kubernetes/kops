package marshaler

import (
	"encoding/json"
	"time"
)

// Duration implements a JSON Marshaler to encode a time.Duration in milliseconds.
type Duration int64

const milliSec = Duration(time.Millisecond)

// NewDuration converts a *time.Duration to a *Duration type.
func NewDuration(t *time.Duration) *Duration {
	if t == nil {
		return nil
	}
	d := Duration(t.Nanoseconds())
	return &d
}

// Standard converts a *Duration to a *time.Duration type.
func (d *Duration) Standard() *time.Duration {
	return (*time.Duration)(d)
}

// MarshalJSON encodes the Duration in milliseconds.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(d / milliSec))
}

// UnmarshalJSON decodes milliseconds to Duration.
func (d *Duration) UnmarshalJSON(b []byte) error {
	var tmp int64
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}
	*d = Duration(tmp) * milliSec
	return nil
}

// DurationSlice is a slice of *Duration
type DurationSlice []*Duration

// NewDurationSlice converts a []*time.Duration to a DurationSlice type.
func NewDurationSlice(t []*time.Duration) DurationSlice {
	ds := make([]*Duration, len(t))
	for i := range ds {
		ds[i] = NewDuration(t[i])
	}
	return ds
}

// Standard converts a DurationSlice to a []*time.Duration type.
func (ds *DurationSlice) Standard() []*time.Duration {
	t := make([]*time.Duration, len(*ds))
	for i := range t {
		t[i] = (*ds)[i].Standard()
	}
	return t
}

// Durationint32Map is a int32 map of *Duration
type Durationint32Map map[int32]*Duration

// NewDurationint32Map converts a map[int32]*time.Duration to a Durationint32Map type.
func NewDurationint32Map(t map[int32]*time.Duration) Durationint32Map {
	dm := make(Durationint32Map, len(t))
	for i := range t {
		dm[i] = NewDuration(t[i])
	}
	return dm
}

// Standard converts a Durationint32Map to a map[int32]*time.Duration type.
func (dm *Durationint32Map) Standard() map[int32]*time.Duration {
	t := make(map[int32]*time.Duration, len(*dm))
	for key, value := range *dm {
		t[key] = value.Standard()
	}
	return t
}

// LongDuration implements a JSON Marshaler to encode a time.Duration in days.
type LongDuration int64

const day = LongDuration(time.Hour) * 24

// NewLongDuration converts a *time.Duration to a *LongDuration type.
func NewLongDuration(t *time.Duration) *LongDuration {
	if t == nil {
		return nil
	}
	d := LongDuration(t.Nanoseconds())
	return &d
}

// Standard converts a *LongDuration to a *time.Duration type.
func (d *LongDuration) Standard() *time.Duration {
	return (*time.Duration)(d)
}

// MarshalJSON encodes the LongDuration in days.
func (d LongDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(d / day))
}

// UnmarshalJSON decodes days to LongDuration.
func (d *LongDuration) UnmarshalJSON(b []byte) error {
	var tmp int64
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}
	*d = LongDuration(tmp) * day
	return nil
}

// LongDurationSlice is a slice of *LongDuration
type LongDurationSlice []*LongDuration

// NewLongDurationSlice converts a []*time.Duration to a LongDurationSlice type.
func NewLongDurationSlice(t []*time.Duration) LongDurationSlice {
	ds := make([]*LongDuration, len(t))
	for i := range ds {
		ds[i] = NewLongDuration(t[i])
	}
	return ds
}

// Standard converts a LongDurationSlice to a []*time.Duration type.
func (ds *LongDurationSlice) Standard() []*time.Duration {
	t := make([]*time.Duration, len(*ds))
	for i := range t {
		t[i] = (*ds)[i].Standard()
	}
	return t
}

// LongDurationint32Map is a int32 map of *LongDuration
type LongDurationint32Map map[int32]*LongDuration

// NewLongDurationint32Map converts a map[int32]*time.LongDuration to a LongDurationint32Map type.
func NewLongDurationint32Map(t map[int32]*time.Duration) LongDurationint32Map {
	dm := make(LongDurationint32Map, len(t))
	for i := range t {
		dm[i] = NewLongDuration(t[i])
	}
	return dm
}

// Standard converts a LongDurationint32Map to a map[int32]*time.LongDuration type.
func (dm *LongDurationint32Map) Standard() map[int32]*time.Duration {
	t := make(map[int32]*time.Duration, len(*dm))
	for key, value := range *dm {
		t[key] = value.Standard()
	}
	return t
}
