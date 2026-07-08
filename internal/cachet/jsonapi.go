package cachet

import (
	"encoding/json"
	"strconv"
)

// The Cachet v3 API speaks JSON:API. These are the minimal envelope types we
// decode against: a document with a data payload (single object or array) and
// an optional side-loaded `included` set.

type document struct {
	Data     json.RawMessage `json:"data"`
	Included []resource      `json:"included"`
}

type resource struct {
	ID            string                  `json:"id"`
	Type          string                  `json:"type"`
	Attributes    json.RawMessage         `json:"attributes"`
	Relationships map[string]relationship `json:"relationships"`
}

// relationship holds a to-one or to-many linkage. Data is a single ref object
// for to-one and an array for to-many, so we keep it raw and decode on demand.
type relationship struct {
	Data json.RawMessage `json:"data"`
}

type resourceRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// one decodes a to-one relationship. ok is false when the link is null/absent.
func (r relationship) one() (ref resourceRef, ok bool) {
	if len(r.Data) == 0 || string(r.Data) == "null" {
		return resourceRef{}, false
	}
	if err := json.Unmarshal(r.Data, &ref); err != nil || ref.ID == "" {
		return resourceRef{}, false
	}
	return ref, true
}

// many decodes a to-many relationship into its member IDs.
func (r relationship) many() []int {
	if len(r.Data) == 0 || string(r.Data) == "null" {
		return nil
	}
	var refs []resourceRef
	if err := json.Unmarshal(r.Data, &refs); err != nil {
		return nil
	}
	out := make([]int, 0, len(refs))
	for _, ref := range refs {
		if id, err := strconv.Atoi(ref.ID); err == nil {
			out = append(out, id)
		}
	}
	return out
}

// Status is Cachet's {human, value} status pair. Its meaning depends on
// context: for components 1=Operational..4=Major Outage; for incidents
// 1=Investigating..4=Fixed.
type Status struct {
	Human string `json:"human"`
	Value int    `json:"value"`
}

// stamp is Cachet's {human, string} timestamp pair. We only keep the machine
// string; parsing to a time.Time happens via LocalTime.
type stamp struct {
	String string `json:"string"`
}
