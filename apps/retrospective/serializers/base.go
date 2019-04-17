package serializers

// BaseRating is the base serializer for rating
type BaseRating struct {
	Rating *int8 `json:"Rating" binding:"omitempty,is_valid_rating"`
}
// BaseResolution is the base serializer for Resolution
type BaseResolution struct {
	Resolution *int8 `json:"Resolution" binding:"omitempty"`
}
