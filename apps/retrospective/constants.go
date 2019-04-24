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

// Resolution ...
type Resolution int8

// ResolutionValues ...
var ResolutionValues = [...]string{
	"-",
	"Done",
	"Won't Do",
	"Duplicate",
	"Can't Reproduce",
}

// GetStringValue ...
func (resolution Resolution) GetStringValue() string {
	return ResolutionValues[resolution]
}

//Resolution
const (
	TaskNotDoneResolution Resolution = iota
	DoneResolution
	WontDoResolution
	DuplicateResolution
	CantReproduceResolution
)
