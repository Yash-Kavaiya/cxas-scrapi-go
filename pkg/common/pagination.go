package common

// PageFunc is called with a page token and returns items plus the next page token.
// Return an empty nextToken to signal the last page.
type PageFunc[T any] func(pageToken string) (items []T, nextToken string, err error)

// Paginate calls fn repeatedly until no more pages remain, collecting all items.
func Paginate[T any](fn PageFunc[T]) ([]T, error) {
	var all []T
	token := ""
	for {
		items, nextToken, err := fn(token)
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
		if nextToken == "" {
			return all, nil
		}
		token = nextToken
	}
}
