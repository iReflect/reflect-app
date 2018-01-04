package serializers

// CategoryDetailSerializer returns the list of questions for a skill of a category
type CategoryDetailSerializer struct {
	ID          uint
	Title       string
	Description string
	Skills      map[uint]SkillDetailSerializer
}
