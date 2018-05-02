package serializers

// BaseRating is the base serializer for rating
type BaseRating struct {
	Rating *int8 `json:"Rating" binding:"omitempty,is_valid_rating"`
}
