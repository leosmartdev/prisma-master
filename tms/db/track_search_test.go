package db

import "testing"
import . "prisma/tms"

func NewTrack(id string, name string) *Track {
	return &Track{
		Id: id,
		Metadata: []*TrackMetadata{
			&TrackMetadata{Name: name},
		},
	}
}

var (
	alpha   = NewTrack("1", "alpha")
	bravo   = NewTrack("2", "bravo")
	charlie = NewTrack("3", "charlie")
	delta   = NewTrack("4", "delta")
)

func TestSorted(t *testing.T) {
	tl := NewTrackLimiter(3)
	tl.Add(charlie)
	tl.Add(bravo)
	tl.Add(alpha)

	results := tl.GetTracks()

	if len(results) != 3 {
		t.Error("Expected length of 3, got", len(results))
	}
	name0 := results[0].Metadata[0].Name
	if name0 != "alpha" {
		t.Error("Expected alpha, got", name0)
	}
	name1 := results[1].Metadata[0].Name
	if name1 != "bravo" {
		t.Error("Expected bravo, got", name1)
	}
	name2 := results[2].Metadata[0].Name
	if name2 != "charlie" {
		t.Error("Expected charlie, got", name2)
	}
}

func TestDiscarded(t *testing.T) {
	tl := NewTrackLimiter(3)
	tl.Add(charlie)
	tl.Add(bravo)
	tl.Add(delta)
	tl.Add(alpha)

	results := tl.GetTracks()
	if len(results) != 3 {
		t.Error("Expected length of 3, got", len(results))
	}
	name2 := results[2].Metadata[0].Name
	if name2 != "charlie" {
		t.Error("Expected charlie, got", name2)
	}
}

func TestReplace(t *testing.T) {
	tl := NewTrackLimiter(3)
	tl.Add(charlie)
	tl.Add(charlie)

	results := tl.GetTracks()
	if len(results) != 1 {
		t.Error("Expected length of 1, got", len(results))
	}
}
