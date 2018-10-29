package ipfs

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

type event struct {
	Time   int64
	Status string
	Node   NodeInfo
}

// Watch listens for specific container events
func (c *client) Watch(ctx context.Context) (<-chan event, <-chan error) {
	var (
		events = make(chan event)
		errs   = make(chan error)
	)

	go func() {
		defer close(errs)
		eventsCh, eventsErrCh := c.d.Events(ctx,
			types.EventsOptions{Filters: filters.NewArgs(
				filters.KeyValuePair{Key: "event", Value: "die"},
				filters.KeyValuePair{Key: "event", Value: "start"},
			)})

		for {
			select {
			case <-ctx.Done():
				break

			// pipe errors back
			case err := <-eventsErrCh:
				if err != nil {
					errs <- err
				}

			// report events
			case status := <-eventsCh:
				id := status.ID[:11]
				name := status.Actor.Attributes["name"]
				node, err := newNode(id, name, status.Actor.Attributes)
				if err != nil {
					continue
				}
				e := event{Time: status.Time, Status: status.Status, Node: node}
				c.l.Infow("event received",
					"event", e)
				events <- e
			}
		}
	}()

	return events, errs
}
