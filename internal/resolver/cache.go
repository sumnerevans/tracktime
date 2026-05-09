package resolver

import (
	"context"
	"encoding/csv"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/timeentry"
)

type cacheKey struct {
	Type    string
	Project string
	TaskID  string
}

type cacheEntry struct {
	Description string
	FetchedAt   time.Time
}

// ItemDetailCache wraps a set of ItemDetailResolvers and provides a persistent
// CSV-backed description cache with soft TTL semantics: stale entries are
// refreshed on access but never deleted on failure, so descriptions survive
// loss of API access.
type ItemDetailCache struct {
	dir       string
	ttl       time.Duration
	resolvers []ItemDetailResolver
	mu        sync.Mutex
	entries   map[cacheKey]cacheEntry
}

// NewItemDetailCache creates an ItemDetailCache, loads the persistent CSV from
// dir, and initialises all resolvers with cfg.
func NewItemDetailCache(dir string, cfg *config.Config, resolvers []ItemDetailResolver) *ItemDetailCache {
	c := &ItemDetailCache{
		dir:       dir,
		ttl:       cfg.CacheTTL(),
		resolvers: resolvers,
		entries:   make(map[cacheKey]cacheEntry),
	}
	for _, r := range resolvers {
		r.Init(cfg)
	}
	c.load()
	return c
}

func (c *ItemDetailCache) cacheFilePath() string {
	return filepath.Join(c.dir, "item-cache.csv")
}

func (c *ItemDetailCache) load() {
	f, err := os.Open(c.cacheFilePath())
	if err != nil {
		return
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil || len(records) == 0 {
		return
	}

	for _, record := range records[1:] { // skip header
		if len(record) < 5 {
			continue
		}
		fetchedAt, err := time.Parse(time.RFC3339, record[4])
		if err != nil {
			continue
		}
		key := cacheKey{Type: record[0], Project: record[1], TaskID: record[2]}
		c.entries[key] = cacheEntry{Description: record[3], FetchedAt: fetchedAt}
	}
}

func (c *ItemDetailCache) save(ctx context.Context) {
	log := zerolog.Ctx(ctx)

	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		log.Error().Err(err).Msg("item cache: mkdir failed")
		return
	}

	tmp, err := os.CreateTemp(c.dir, ".item-cache-*.csv")
	if err != nil {
		log.Error().Err(err).Msg("item cache: failed to create temp file")
		return
	}
	tmpPath := tmp.Name()

	w := csv.NewWriter(tmp)
	_ = w.Write([]string{"type", "project", "taskid", "description", "fetched_at"})
	for key, entry := range c.entries {
		_ = w.Write([]string{
			key.Type, key.Project, key.TaskID,
			entry.Description,
			entry.FetchedAt.UTC().Format(time.RFC3339),
		})
	}
	w.Flush()
	tmp.Close()

	if err := os.Rename(tmpPath, c.cacheFilePath()); err != nil {
		log.Error().Err(err).Msg("item cache: failed to save")
		os.Remove(tmpPath)
	}
}

// GetFormattedTaskID returns the service-specific formatted task ID (e.g. "#123",
// "ENG-456"). Falls back to the raw TaskID string if no resolver handles the entry.
func (c *ItemDetailCache) GetFormattedTaskID(entry *timeentry.TimeEntry) string {
	for _, r := range c.resolvers {
		if r.Handles(entry) {
			if id := r.GetFormattedTaskID(entry); id != "" {
				return id
			}
		}
	}
	return string(entry.TaskID)
}

// GetTaskLink returns the URL for the work item, or "" if unknown.
func (c *ItemDetailCache) GetTaskLink(entry *timeentry.TimeEntry) string {
	for _, r := range c.resolvers {
		if r.Handles(entry) {
			return r.GetTaskLink(entry)
		}
	}
	return ""
}

// GetDescription returns the human-readable title/description for the work
// item. It checks the cache first; stale entries trigger a re-fetch but are
// returned as-is if the fetch fails.
func (c *ItemDetailCache) GetDescription(ctx context.Context, entry *timeentry.TimeEntry) string {
	log := zerolog.Ctx(ctx)

	if entry.TaskID == "" {
		return ""
	}

	key := cacheKey{
		Type:    string(entry.Type),
		Project: string(entry.Project),
		TaskID:  string(entry.TaskID),
	}

	c.mu.Lock()
	cached, exists := c.entries[key]
	c.mu.Unlock()

	now := time.Now()
	if exists && now.Sub(cached.FetchedAt) <= c.ttl {
		return cached.Description
	}

	// Stale or missing — attempt fetch.
	var desc string
	for _, r := range c.resolvers {
		if r.Handles(entry) {
			fetched, err := r.FetchDescription(ctx, entry)
			if err != nil {
				if exists {
					log.Warn().Err(err).
						Str("type", key.Type).
						Str("project", key.Project).
						Str("taskid", key.TaskID).
						Msg("item cache: refresh failed, keeping stale entry")
					return cached.Description
				}
				return ""
			}
			desc = fetched
			break
		}
	}

	if desc == "" && exists {
		return cached.Description
	}

	if desc != "" {
		c.mu.Lock()
		c.entries[key] = cacheEntry{Description: desc, FetchedAt: now}
		c.mu.Unlock()
		c.save(ctx)
	}

	return desc
}
