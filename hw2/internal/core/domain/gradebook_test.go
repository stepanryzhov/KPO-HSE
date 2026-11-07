package domain

import "testing"

func testGradebook(t *testing.T) (*Gradebook, studentID, []Criterion) {
	t.Helper()
	registry, err := NewStudentRegistry(1)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.RestoreStudent(1, "Анна"); err != nil {
		t.Fatal(err)
	}
	sid := registry.Students()[0].GetID()
	book := NewGradebook()
	if err := registry.AddToGradebook(book, sid); err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct {
		name   string
		weight float64
	}{{"ДЗ", 0.25}, {"Экзамен", 0.75}} {
		criterionName, err := NewName(item.name)
		if err != nil {
			t.Fatal(err)
		}
		maxScore, err := NewScore(10)
		if err != nil {
			t.Fatal(err)
		}
		criterionWeight, err := NewWeight(item.weight)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := book.CreateCriterion(criterionName, maxScore, criterionWeight); err != nil {
			t.Fatal(err)
		}
	}
	return book, sid, book.Criteria()
}

func TestEvaluationScoreLimits(t *testing.T) {
	for _, tc := range []struct {
		name      string
		value     float64
		wantError bool
	}{
		{"отрицательный балл", -1, true},
		{"нулевой балл", 0, false},
		{"максимальный балл", 10, false},
		{"выше максимума", 11, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			book, sid, criteria := testGradebook(t)
			value, err := NewScore(tc.value)
			if err == nil {
				err = book.SetEvaluation(sid, criteria[0].GetID(), value)
			}
			if (err != nil) != tc.wantError {
				t.Fatalf("балл %v: ошибка = %v, ожидалась ошибка: %v", tc.value, err, tc.wantError)
			}
		})
	}
}

func TestEvaluationRequiresExistingCriterion(t *testing.T) {
	book, sid, _ := testGradebook(t)
	unknownID, err := NewID(999)
	if err != nil {
		t.Fatal(err)
	}
	value, err := NewScore(5)
	if err != nil {
		t.Fatal(err)
	}
	if err := book.SetEvaluation(sid, unknownID, value); err == nil {
		t.Fatal("нельзя выставить оценку по неизвестному критерию")
	}
}

func TestCalculateFinalMark(t *testing.T) {
	book, sid, criteria := testGradebook(t)
	for i, rawScore := range []float64{8, 10} {
		value, err := NewScore(rawScore)
		if err != nil {
			t.Fatal(err)
		}
		if err := book.SetEvaluation(sid, criteria[i].GetID(), value); err != nil {
			t.Fatal(err)
		}
	}
	mark, err := CalculateFinalMark(book, sid)
	if err != nil {
		t.Fatal(err)
	}
	if mark.Value() != 9.5 {
		t.Fatalf("итог = %v, ожидается 9.5", mark.Value())
	}
}

func TestFinalMarkAfterEvaluationChange(t *testing.T) {
	book, sid, criteria := testGradebook(t)
	value, err := NewScore(8)
	if err != nil {
		t.Fatal(err)
	}
	for _, criterion := range criteria {
		if err := book.SetEvaluation(sid, criterion.GetID(), value); err != nil {
			t.Fatal(err)
		}
	}
	mark, err := CalculateFinalMark(book, sid)
	if err != nil || mark.Value() != 8 {
		t.Fatalf("исходный итог = %v, ошибка = %v; ожидается 8", mark.Value(), err)
	}
	corrected, err := NewScore(10)
	if err != nil {
		t.Fatal(err)
	}
	if err := book.SetEvaluation(sid, criteria[0].GetID(), corrected); err != nil {
		t.Fatal(err)
	}
	mark, err = CalculateFinalMark(book, sid)
	if err != nil || mark.Value() != 8.5 {
		t.Fatalf("итог после исправления = %v, ошибка = %v; ожидается 8.5", mark.Value(), err)
	}
	if err := book.RemoveEvaluation(sid, criteria[0].GetID()); err != nil {
		t.Fatal(err)
	}
	if _, err := CalculateFinalMark(book, sid); err == nil {
		t.Fatal("после отмены оценки итог нельзя рассчитать: не хватает оценки по критерию")
	}
}
