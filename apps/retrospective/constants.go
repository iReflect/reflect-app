package retrospective

// RatingValues ...
var RatingValues = [...]string{
	"Ugly",
	"Bad",
	"Okay",
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
	UglyRating Rating = iota
	BadRating
	OkayRating
	GoodRating
	NotableRating
)
