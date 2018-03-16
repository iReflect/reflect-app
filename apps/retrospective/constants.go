package retrospective

// RatingValues ...
var RatingValues = [...]string{
	"Red",
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
	RedRating Rating = iota
	ImproveRating
	DecentRating
	GoodRating
	NotableRating
)
