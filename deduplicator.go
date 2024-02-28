package htadaptor

import (
	"context"
	"sync"
	"time"

	"github.com/mitchellh/hashstructure/v2"
)

// Deduplicator tracks recent request structures.
// Use [Deduplicator.IsDuplicate] to determine if a request
// was seen in a certain time window.
//
// Uses struct hashing by Mitchell Hashimoto. Using request
// deduplicator is superior to deduplicating HTTP requests,
// because they are large and can vary in myriads of ways.
// Struct deduplication can also be applied at network edge
// easing the pressure on the database or the event bus.
//
// Handles most data structs. Cannot process functions
// inside structs.
type Deduplicator struct {
	window time.Duration
	mu     *sync.Mutex
	tags   map[uint64]time.Time
}

func (d *Deduplicator) cleanOutLoop(ctx context.Context, ticker *time.Ticker) {
	for {
		select {
		case <-ctx.Done():
			return // execution ended, part the go routine
		case tagsBefore := <-ticker.C:
			d.cleanOut(tagsBefore)
		}
	}
}

func (d *Deduplicator) cleanOut(tagsBefore time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for hash, expires := range d.tags {
		if expires.Before(tagsBefore) {
			delete(d.tags, hash)
		}
	}
}

// Len returns the number of known tags that have not been
// cleaned out yet.
func (d *Deduplicator) Len() (count int) {
	d.mu.Lock()
	count = len(d.tags)
	d.mu.Unlock()
	return
}

// IsDuplicate returns true if the message hash tag calculated
// using a [MessageHasher] was seen in deduplication time window.
func (d *Deduplicator) IsDuplicate(request any) (bool, error) {
	tag, err := hashstructure.Hash(request, hashstructure.FormatV2, nil)
	if err != nil {
		return false, &DecodingError{error: err}
	}

	d.mu.Lock()
	_, alreadySeen := d.tags[tag]
	if alreadySeen {
		// NOTE: could also check if tag expires.After(t)
		// and remove it for exact expiration
		// instead of fuzzy until-next clean up expiration
		// but this should not be needed for most use cases.
		d.mu.Unlock()
		return true, nil
	}
	d.tags[tag] = time.Now().Add(d.window)
	d.mu.Unlock()
	return false, nil // first time, not a duplicate
}

// NewDeduplicator returns a new Deduplicator.
//
// Context is used for the clean up go routine termination.
//
// Window specifies the minimum duration of how long the
// duplicate tags are remembered for. Real duration can
// extend up to 50% longer because it depends on the
// clean up cycle.
func NewDeduplicator(ctx context.Context, window time.Duration) *Deduplicator {
	if window < time.Millisecond {
		panic("deduplication window of less than a millisecond is impractical")
	}

	d := &Deduplicator{
		window: window,

		mu:   &sync.Mutex{},
		tags: make(map[uint64]time.Time),
	}
	go d.cleanOutLoop(ctx, time.NewTicker(window/2))
	return d
}
