// Package sel provides helpers for different selection techniques.
//
// Selection helpers come in two varieties, function selectors and pool
// selectors. Function selectors are simply functions that take some number of
// genomes as "competitors" and return one or more "winners" from the input.
//
// Pool selectors allow many goroutines to contribute competitors and for the
// winners to be retrieved individually. Once all the winners are retrieved,
// the pool is reset for another round of competition.
package sel
