package retrospective

// RatingValues ...
var RatingValues = [...]string{
	"Concern",
	"Improve",
	"Decent",
	"Good",
	"Notable",
}

// Rating ...
type Rating int8

// GetStringValue ...
func (rating Rating) GetStringValue() string {
	return RatingValues[rating]
}

// Rating
const (
	ConcernRating Rating = iota
	ImproveRating
	DecentRating
	GoodRating
	NotableRating
)
