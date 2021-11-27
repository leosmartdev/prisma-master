// Package omngen contains extra function to generate payloads or handle beacon behavior.
package omngen

import "prisma/tms/omnicom"
import "time"
import "github.com/StefanSchroeder/Golang-Ellipsoid/ellipsoid"

func DestSwap(a, b Pos) (Pos, Pos) {
	return b, a
}

//increment time by the reporting period and returns it in the omnicom format
func IncTime(date omnicom.Date_Position, i int) (omnicom.Date_Position, error) {

	var t time.Time
	t = time.Date((2000 + int(date.Year)), time.Month(date.Month), int(date.Day), 0, int(date.Minute)+i, 0, 0, time.UTC)
	date.Year = uint32(t.Year() - 2000)
	date.Month = uint32(t.Month())
	date.Day = uint32(t.Day())
	date.Minute = uint32((uint32(t.Hour()) * 60) + uint32(t.Minute()))

	return date, nil
}

//CurrentTime reflects current date to omnicom.Dateposition
func CurrentTime() omnicom.Date_Position {

	var date omnicom.Date_Position
	date.Year = uint32(time.Now().Year() - 2000)
	date.Month = uint32(time.Now().Month())
	date.Day = uint32(time.Now().Day())
	date.Minute = uint32((uint32(time.Now().Hour()) * 60) + uint32(time.Now().Minute()))

	return date
}

//OmnicomDate reflects current date to omnicom.Date
func OmnicomDate() omnicom.Date {

	var date omnicom.Date
	date.Year = uint32(time.Now().Year() - 2000)
	date.Month = uint32(time.Now().Month())
	date.Day = uint32(time.Now().Day())
	date.Minute = uint32((uint32(time.Now().Hour()) * 60) + uint32(time.Now().Minute()))

	return date

}

//Move moves the vessel form current position and returns next position
func Move(lat, long float64, step, bearing float64, geo ellipsoid.Ellipsoid) (Pos, error) {
	var p Pos
	p.Lat, p.Long = geo.At(lat, long, step, bearing)

	return p, nil
}
