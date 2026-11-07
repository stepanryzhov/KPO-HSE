package domain

type Evaluation struct {
	studentID   studentID
	criterionID criterionID
	score       score
}

func NewEvaluation(student Student, criterion Criterion, value score) (Evaluation, error) {
	if err := student.validate(); err != nil {
		return Evaluation{}, err
	}
	if err := criterion.validate(); err != nil {
		return Evaluation{}, err
	}

	studentID := student.GetID()
	criterionID := criterion.GetID()

	if err := criterion.ValidateScore(value); err != nil {
		return Evaluation{}, err
	}

	return Evaluation{
		studentID:   studentID,
		criterionID: criterionID,
		score:       value,
	}, nil
}

func (e Evaluation) GetStudentID() studentID {
	return e.studentID
}

func (e Evaluation) GetCriterionID() criterionID {
	return e.criterionID
}

func (e Evaluation) GetScore() score {
	return e.score
}
