package ledger

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrAccountClosed is returned when a posting targets a closed account.
var ErrAccountClosed = errors.New("ledger: account is closed")

// Entry is a single double-entry posting against one account.
type Entry struct {
	AccountID string
	Amount    int64
	Currency  string
	PostedAt  time.Time
	Memo      string
}

// Journal collects entries until they balance, then commits them together.
type Journal struct {
	mu      sync.Mutex
	pending []Entry
	clock   func() time.Time
	sink    Sink
}

// Sink persists a balanced batch of entries.
type Sink interface {
	Commit(ctx context.Context, batch []Entry) error
}

// NewJournal returns a journal that writes balanced batches to sink.
func NewJournal(sink Sink) *Journal {
	return &Journal{clock: time.Now, sink: sink}
}

// Post appends an entry to the pending batch. It does not persist anything.
func (j *Journal) Post(accountID string, amount int64, currency, memo string) error {
	if accountID == "" {
		return fmt.Errorf("ledger: empty account id")
	}
	if currency == "" {
		return fmt.Errorf("ledger: entry for %s has no currency", accountID)
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	j.pending = append(j.pending, Entry{
		AccountID: accountID,
		Amount:    amount,
		Currency:  currency,
		PostedAt:  j.clock(),
		Memo:      memo,
	})
	return nil
}

// Balance reports the net amount per currency across the pending batch.
func (j *Journal) Balance() map[string]int64 {
	j.mu.Lock()
	defer j.mu.Unlock()
	out := make(map[string]int64, 4)
	for _, e := range j.pending {
		out[e.Currency] += e.Amount
	}
	return out
}

// Flush commits the pending batch if every currency nets to zero.
func (j *Journal) Flush(ctx context.Context) error {
	j.mu.Lock()
	batch := j.pending
	j.pending = nil
	j.mu.Unlock()

	if len(batch) == 0 {
		return nil
	}
	net := make(map[string]int64, 4)
	for _, e := range batch {
		net[e.Currency] += e.Amount
	}
	for cur, amount := range net {
		if amount != 0 {
			return fmt.Errorf("ledger: batch does not balance in %s: net %d", cur, amount)
		}
	}
	if err := j.sink.Commit(ctx, batch); err != nil {
		return fmt.Errorf("ledger: commit %d entries: %w", len(batch), err)
	}
	return nil
}
