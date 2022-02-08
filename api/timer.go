package api

import (
	"fmt"
	"log"
	"time"
)

func (api *API) Tick(now time.Time) error {
	datasets, err := api.ReadDatasets()
	if err != nil {
		return err
	}

	hash, err := api.LatestConfigHash()
	if err != nil {
		return fmt.Errorf("read latest config hash: %w", err)
	}

	// refresh datasets
	for _, dataset := range datasets {
		if dataset.Refresh != nil {
			dur, err := time.ParseDuration(dataset.Refresh.Interval)
			if err != nil {
				return fmt.Errorf("parse refresh interval: %w", err)
			}

			if lastRefresh, ok := api.lastRefresh.Load(dataset.ID); !ok || lastRefresh.(time.Time).Before(now.Add(-dur)) {
				log.Println("refreshing", dataset.ID)
				err = api.refreshDataset(hash, dataset.ID)
				if err != nil {
					return fmt.Errorf("refreshing %s: %w", dataset.ID, err)
				}
				api.lastRefresh.Store(dataset.ID, now)
			}
		}
	}

	return nil
}
