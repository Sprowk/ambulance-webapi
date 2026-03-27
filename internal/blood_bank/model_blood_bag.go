package blood_bank

import "time"

type BloodBag struct {
	Id             string    `json:"id"`
	BloodGroup     string    `json:"bloodGroup"`
	RhFactor       string    `json:"rhFactor"`
	CollectionDate time.Time `json:"collectionDate"`
	Volume         int32     `json:"volume"`
	Status         string    `json:"status"`
	DonorId        string    `json:"donorId,omitempty"`
	Notes          string    `json:"notes,omitempty"`
	CreatedAt      time.Time `json:"createdAt,omitempty"`
}
