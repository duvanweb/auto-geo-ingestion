package resources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"auto-geo-ingestion/internal/core/domain"
	portresources "auto-geo-ingestion/internal/core/ports/resources"

	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

type locationCache struct {
	client *redis.Client
}

// Dependencies holds the dependencies for the location cache resource.
type Dependencies struct {
	fx.In
	Client *redis.Client
}

// NewLocationCache creates a new LocationCache backed by Redis.
func NewLocationCache(deps Dependencies) portresources.LocationCache {
	return &locationCache{client: deps.Client}
}

// DeleteVehicleKeys removes all cached keys associated with a vehicle.
func (c *locationCache) DeleteVehicleKeys(ctx context.Context, vehicleID uint64) error {
	pattern := fmt.Sprintf("location:vehicle:%d:*", vehicleID)

	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("scanning cache keys for vehicle %d: %w", vehicleID, err)
	}

	if len(keys) == 0 {
		return nil
	}

	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("deleting cache keys for vehicle %d: %w", vehicleID, err)
	}

	return nil
}

// GetLastLocation retrieves the last known location for a vehicle from Redis.
// Returns false as the second value when no entry exists in the cache.
func (c *locationCache) GetLastLocation(ctx context.Context, vehicleID uint64) (domain.Location, bool, error) {
	key := lastLocationKey(vehicleID)

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return domain.Location{}, false, nil
		}

		return domain.Location{}, false, fmt.Errorf("getting last location for vehicle %d: %w", vehicleID, err)
	}

	var loc domain.Location
	if err := json.Unmarshal(data, &loc); err != nil {
		return domain.Location{}, false, fmt.Errorf("unmarshalling last location for vehicle %d: %w", vehicleID, err)
	}

	return loc, true, nil
}

// SetLastLocation stores the last known location for a vehicle with the given TTL.
func (c *locationCache) SetLastLocation(ctx context.Context, loc domain.Location, ttl time.Duration) error {
	data, err := json.Marshal(loc)
	if err != nil {
		return fmt.Errorf("marshalling location for vehicle %d: %w", loc.VehicleID, err)
	}

	if err := c.client.Set(ctx, lastLocationKey(loc.VehicleID), data, ttl).Err(); err != nil {
		return fmt.Errorf("setting last location for vehicle %d: %w", loc.VehicleID, err)
	}

	return nil
}

func lastLocationKey(vehicleID uint64) string {
	return fmt.Sprintf("location:vehicle:%d:last", vehicleID)
}
