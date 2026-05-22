package akahu

// Cursor is the pagination cursor returned by paginated list endpoints.
// A nil pointer means there are no more pages.
type Cursor struct {
	Next *string `json:"next"`
}

// Page is a single page of results from a paginated endpoint. To fetch the
// next page, pass Cursor.Next as the cursor query parameter on the next call.
// When Cursor.Next is nil, the last page has been reached.
type Page[T any] struct {
	Items  []T    `json:"items"`
	Cursor Cursor `json:"cursor"`
}
