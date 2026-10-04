package score

import "time"

func newEntryFixture() Entry {
	return Entry{
		Kills:    5,
		Duds:     1,
		FreedMem: 1024,
		Speed:    2.0,
		Time:     60,
		Duration: 10.0,
		Date:     time.Date(2026, time.October, 4, 12, 0, 0, 0, time.UTC),
	}
}

func newEntryFixtureFor(kills int) Entry {
	entry := newEntryFixture()
	entry.Kills = kills
	return entry
}
