package serializers

// SkillDetailSerializer returns the skill details with a list of questions under that skill
type SkillDetailSerializer struct {
	ID           uint
	Title        string
	DisplayTitle string
	Description  string
	Weight       int
	Questions    []QuestionResponseDetailSerializer
}

