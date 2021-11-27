package db

// For paged requests, set the page number to retrieve and the length of
// each page. To simply limit the results, set the length and Number can
// be zero. Use an zeroed struct to get all results.
type Page struct {
	Number int
	Length int
}
