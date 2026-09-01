package repositories

import (
	"context"
	"sync"
	"time"

	"auto-geo-ingestion/internal/core/domain"
	portrepos "auto-geo-ingestion/internal/core/ports/repositories"
	"auto-geo-ingestion/internal/infrastructure/pkg/logger"

	"github.com/sony/gobreaker"
)

const fallbackQueueSize = 1000

// LocationRepositoryWithCB wraps LocationRepository with a circuit breaker and an in-memory
// fallback queue so that transient Postgres failures do not crash the ingestion pipeline.
type LocationRepositoryWithCB struct {
	inner    *LocationRepository
	cb       *gobreaker.CircuitBreaker
	fallback chan domain.Location
	log      logger.Logger
	once     sync.Once
}

// NewLocationRepositoryWithCB creates a LocationRepository decorator that protects the underlying
// Postgres repository with a circuit breaker and a bounded in-memory fallback queue.
func NewLocationRepositoryWithCB(inner *LocationRepository, log logger.Logger) portrepos.LocationRepository {
	settings := gobreaker.Settings{
		Name:        "postgres-location",
		MaxRequests: 3,
		Interval:    30 * time.Second,
		Timeout:     60 * time.Second,
		ReadyToTrip: func(c gobreaker.Counts) bool {
			return c.ConsecutiveFailures > 3
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Errorw(context.Background(), "circuit breaker state change",
				"name", name, "from", from.String(), "to", to.String())
		},
	}

	return &LocationRepositoryWithCB{
		inner:    inner,
		cb:       gobreaker.NewCircuitBreaker(settings),
		fallback: make(chan domain.Location, fallbackQueueSize),
		log:      log,
	}
}

// FindLatestByVehicle delegates to the inner repository through the circuit breaker.
func (r *LocationRepositoryWithCB) FindLatestByVehicle(ctx context.Context, vehicleID uint64, limit int) ([]domain.Location, error) {
	result, err := r.cb.Execute(func() (any, error) {
		return r.inner.FindLatestByVehicle(ctx, vehicleID, limit)
	})
	if err != nil {
		return nil, err
	}

	locs, _ := result.([]domain.Location)

	return locs, nil
}

// Save persists a location through the circuit breaker. When the circuit is open or the DB
// call fails, the location is enqueued in the fallback buffer and nil error is returned so
// the ingestion pipeline can continue.
func (r *LocationRepositoryWithCB) Save(ctx context.Context, loc domain.Location) (*domain.Location, error) {
	result, err := r.cb.Execute(func() (any, error) {
		return r.inner.Save(ctx, loc)
	})
	if err != nil {
		select {
		case r.fallback <- loc:
			r.log.Errorw(ctx, "location queued in fallback (DB circuit open)", "vehicle_id", loc.VehicleID)
		default:
			r.log.Errorw(ctx, "fallback queue full, location discarded", "vehicle_id", loc.VehicleID)
		}

		r.startDrain()

		return nil, nil
	}

	r.startDrain()

	saved, _ := result.(*domain.Location)

	return saved, nil
}

// drainQueue periodically retries queued locations against the inner repository when the
// circuit breaker is closed.
func (r *LocationRepositoryWithCB) drainQueue() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if len(r.fallback) == 0 {
			continue
		}

		if r.cb.State() != gobreaker.StateClosed {
			continue
		}

		for {
			select {
			case loc := <-r.fallback:
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

				if _, err := r.inner.Save(ctx, loc); err != nil {
					cancel()

					select {
					case r.fallback <- loc:
					default:
					}

					goto nextTick
				}

				cancel()

			default:
				goto nextTick
			}
		}

	nextTick:
	}
}

// startDrain ensures the background drain goroutine is started exactly once.
func (r *LocationRepositoryWithCB) startDrain() {
	r.once.Do(func() {
		go r.drainQueue()
	})
}
