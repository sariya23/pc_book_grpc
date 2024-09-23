package storage

import "sync"

type Rating struct {
	Count uint32
	Sum   float64
}

type RatingStorage struct {
	mu     sync.RWMutex
	rating map[string]*Rating
}

func NewRatingStorage() *RatingStorage {
	return &RatingStorage{
		rating: make(map[string]*Rating),
	}
}

func (rs *RatingStorage) Add(laptopId string, score float64) (*Rating, error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rating := rs.rating[laptopId]
	if rating == nil {
		rating = &Rating{
			Count: 1,
			Sum:   score,
		}
	} else {
		rating.Count++
		rating.Sum += score
	}
	rs.rating[laptopId] = rating
	return rating, nil
}
