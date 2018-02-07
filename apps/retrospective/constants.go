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

// String ...
func (rating Rating) String() string {
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
