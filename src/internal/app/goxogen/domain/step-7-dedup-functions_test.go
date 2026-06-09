package domainapp

import (
	"strings"
	"testing"
)

func TestDedup_ConsecutiveTopLevel(t *testing.T) {
	input := `// Generated from index 'search_performerclosedate_index'.
func GetTasksByPerformerCloseDate(db DBQuery, d sql.NullTime) ([]*Task, error) {
	return nil, nil
}

// Generated from index 'task_performer_close_date_index'.
func GetTasksByPerformerCloseDate(db DBQuery, d sql.NullTime) ([]*Task, error) {
	return nil, nil
}

func AnotherFunc() int {
	return 42
}
`
	seen := make(map[string]bool)
	got := removeDuplicateFuncs(input, seen)

	if strings.Count(got, "func GetTasksByPerformerCloseDate(") != 1 {
		t.Fatalf("expected 1 GetTasksByPerformerCloseDate, got:\n%s", got)
	}
	if strings.Count(got, "func AnotherFunc()") != 1 {
		t.Fatalf("expected 1 AnotherFunc, got:\n%s", got)
	}
	if strings.Contains(got, "task_performer_close_date_index") {
		t.Fatalf("second index comment should be removed, got:\n%s", got)
	}
}

func TestDedup_NonConsecutive(t *testing.T) {
	input := `func GetTasksByPerformerCloseDate(db DBQuery, d sql.NullTime) ([]*Task, error) {
	return nil, nil
}

func SomeOtherFunc(x int) int {
	return x * 2
}

// duplicate, not consecutive
func GetTasksByPerformerCloseDate(db DBQuery, d sql.NullTime) ([]*Task, error) {
	return nil, nil
}

func LastFunc() string {
	return "done"
}
`
	seen := make(map[string]bool)
	got := removeDuplicateFuncs(input, seen)

	if strings.Count(got, "func GetTasksByPerformerCloseDate(") != 1 {
		t.Fatalf("expected 1 GetTasksByPerformerCloseDate, got:\n%s", got)
	}
	if strings.Count(got, "func SomeOtherFunc(") != 1 {
		t.Fatalf("expected 1 SomeOtherFunc, got:\n%s", got)
	}
	if strings.Count(got, "func LastFunc(") != 1 {
		t.Fatalf("expected 1 LastFunc, got:\n%s", got)
	}
	// SomeOtherFunc должна остаться между первой и дубликатом
	if !strings.Contains(got, "SomeOtherFunc") {
		t.Fatalf("SomeOtherFunc should survive, got:\n%s", got)
	}
}

func TestDedup_MethodDuplicate(t *testing.T) {
	input := `// FindAll by index 'a'.
func (repo *TaskRepository) FindAllTasksByPerformerCloseDate(d sql.NullTime) ([]*Task, error) {
	return GetTasksByPerformerCloseDate(repo.db, d)
}

// FindAll by index 'b'.
func (repo *TaskRepository) FindAllTasksByPerformerCloseDate(d sql.NullTime) ([]*Task, error) {
	return GetTasksByPerformerCloseDate(repo.db, d)
}

func (repo *TaskRepository) FindAllByStatus(s string) ([]*Task, error) {
	return GetTasksByStatus(repo.db, s)
}
`
	seen := make(map[string]bool)
	got := removeDuplicateFuncs(input, seen)

	if strings.Count(got, "FindAllTasksByPerformerCloseDate") != 1 {
		t.Fatalf("expected 1 FindAllTasksByPerformerCloseDate method, got:\n%s", got)
	}
	if strings.Count(got, "FindAllByStatus") != 1 {
		t.Fatalf("expected 1 FindAllByStatus, got:\n%s", got)
	}
}

func TestDedup_DifferentReceiversSameName(t *testing.T) {
	// Одинаковые имена методов на РАЗНЫХ ресиверах — НЕ дубликаты
	input := `func (repo *TaskRepository) FindAll() ([]*Task, error) {
	return nil, nil
}

func (repo *UserRepository) FindAll() ([]*User, error) {
	return nil, nil
}
`
	seen := make(map[string]bool)
	got := removeDuplicateFuncs(input, seen)

	if strings.Count(got, "func (repo *") != 2 {
		t.Fatalf("expected 2 method declarations (different receivers), got:\n%s", got)
	}
}

func TestDedup_GlobalSeenAcrossFiles(t *testing.T) {
	file1 := `func GetX(db DBQuery, x int) ([]*Task, error) {
	return nil, nil
}
`
	file2 := `func GetX(db DBQuery, x int) ([]*Task, error) {
	return nil, nil
}

func GetY(db DBQuery, y string) ([]*User, error) {
	return nil, nil
}
`
	seen := make(map[string]bool)

	got1 := removeDuplicateFuncs(file1, seen)
	if !strings.Contains(got1, "GetX") {
		t.Fatalf("file1 should keep GetX")
	}

	got2 := removeDuplicateFuncs(file2, seen)
	if strings.Contains(got2, "func GetX(") {
		t.Fatalf("file2 should remove GetX (global duplicate), got:\n%s", got2)
	}
	if !strings.Contains(got2, "func GetY(") {
		t.Fatalf("file2 should keep GetY, got:\n%s", got2)
	}
}

func TestDedup_MultiLineSignature(t *testing.T) {
	input := `func GetTasksByPerformerCloseDate(
	db DBQuery,
	d sql.NullTime,
) ([]*Task, error) {
	return nil, nil
}

func GetTasksByPerformerCloseDate(
	db DBQuery,
	d sql.NullTime,
) ([]*Task, error) {
	return nil, nil
}
`
	seen := make(map[string]bool)
	got := removeDuplicateFuncs(input, seen)

	if strings.Count(got, "func GetTasksByPerformerCloseDate(") != 1 {
		t.Fatalf("expected 1 with multi-line signature, got:\n%s", got)
	}
}

func TestDedup_NoChangesIfNoDupes(t *testing.T) {
	input := `func A() int {
	return 1
}

func B() int {
	return 2
}
`
	seen := make(map[string]bool)
	got := removeDuplicateFuncs(input, seen)

	if got != input {
		t.Fatalf("expected unchanged, got:\n%s", got)
	}
}

func TestDedup_EmptyFile(t *testing.T) {
	seen := make(map[string]bool)
	got := removeDuplicateFuncs("", seen)
	if got != "" {
		t.Fatalf("expected empty, got: %q", got)
	}
}

func TestDedup_NoFuncs(t *testing.T) {
	input := `package main

var x = 42
`
	seen := make(map[string]bool)
	got := removeDuplicateFuncs(input, seen)
	if got != input {
		t.Fatalf("expected unchanged, got: %q", got)
	}
}

func TestDedup_FirstFuncIsDuplicate(t *testing.T) {
	// Первая функция в файле — дубликат, уже виденный в другом файле.
	// package/imports должны сохраниться.
	input := `package mypackage

import "fmt"

func GetX(db DBQuery, x int) ([]*Task, error) {
	return nil, nil
}

func GetY(db DBQuery, y string) ([]*User, error) {
	return nil, nil
}
`
	seen := map[string]bool{"GetX": true} // already seen from another file
	got := removeDuplicateFuncs(input, seen)

	if !strings.Contains(got, "package mypackage") {
		t.Fatalf("package declaration lost, got:\n%s", got)
	}
	if !strings.Contains(got, `import "fmt"`) {
		t.Fatalf("import lost, got:\n%s", got)
	}
	if strings.Contains(got, "func GetX(") {
		t.Fatalf("duplicate GetX should be removed, got:\n%s", got)
	}
	if strings.Count(got, "func GetY(") != 1 {
		t.Fatalf("expected 1 GetY, got:\n%s", got)
	}
}

func TestDedup_ThreeDupesKeepFirst(t *testing.T) {
	input := `func A() int { return 1 }
func A() int { return 2 }
func A() int { return 3 }
func B() int { return 4 }
`
	seen := make(map[string]bool)
	got := removeDuplicateFuncs(input, seen)

	if strings.Count(got, "func A()") != 1 {
		t.Fatalf("expected 1 A(), got:\n%s", got)
	}
	if !strings.Contains(got, "return 1") {
		t.Fatalf("expected first A() body (return 1), got:\n%s", got)
	}
	if strings.Contains(got, "return 2") || strings.Contains(got, "return 3") {
		t.Fatalf("second/third A() bodies should be removed, got:\n%s", got)
	}
	if strings.Count(got, "func B()") != 1 {
		t.Fatalf("expected 1 B(), got:\n%s", got)
	}
}
