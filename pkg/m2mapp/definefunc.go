package m2mapp

import (
	"fmt"
)

type App struct {
	AppID       string
	Address     string
	Description string
	GID         uint64 // goroutine ID
}

type AppInputData struct {
	AppName    string
	NE         SquarePoint
	SW         SquarePoint
	Capability []string
	Period     PeriodInput
	Limit      Range
	Timeout    int
	Action     string
	Parameter  float64
}

type SquarePoint struct {
	Lat float64
	Lon float64
}

type PeriodInput struct {
	Start string
	End   string
}

type Range struct {
	LowerLimit float64
	UpperLimit float64
}

func (a *App) String() string {
	return fmt.Sprintf("AppID: %s, Address: %s, Description: %s, GID: %d", a.AppID, a.Address, a.Description, a.GID)
}
