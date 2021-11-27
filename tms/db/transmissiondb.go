package db

import "prisma/tms"

// TransmissionDB is an interface to work with transmission
type TransmissionDB interface {
	// Create transmission, add Id
	Create(tr *tms.Transmission) error
	// Update whole structure
	Update(tr *tms.Transmission) error
	// Set success status for transmission with specific message Id
	Status(requestId string, state tms.Transmission_State, status int32) error
	// FindByID
	FindByID(id string) (*tms.Transmission, error)
}
