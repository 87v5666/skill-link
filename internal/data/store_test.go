package data

import (
	"os"
	"testing"
)

func TestLoad_Default(t *testing.T) {
	s, err := Load()
	if err != nil {
		t.Fatalf("Load() 不应报错: %v", err)
	}
	if s.Categories == nil {
		t.Fatal("Categories 不应为 nil")
	}
	if s.Notes == nil {
		t.Fatal("Notes 不应为 nil")
	}
}

func TestCreateCategory(t *testing.T) {
	// Use a temp dir to avoid polluting real config
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", t.TempDir())
	defer os.Setenv("HOME", oldHome)

	s, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if err := s.CreateCategory("my-cat"); err != nil {
		t.Fatalf("CreateCategory failed: %v", err)
	}
	if _, ok := s.Categories["my-cat"]; !ok {
		t.Fatal("expected category to exist")
	}
	// Duplicate should error
	if err := s.CreateCategory("my-cat"); err == nil {
		t.Fatal("expected error for duplicate category")
	}
	// Empty name should error
	if err := s.CreateCategory(""); err == nil {
		t.Fatal("expected error for empty category name")
	}
}

func TestAddRemoveSkillFromCategory(t *testing.T) {
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", t.TempDir())
	defer os.Setenv("HOME", oldHome)

	s, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	s.CreateCategory("fe")
	s.AddSkillToCategory("fe", "react")
	s.AddSkillToCategory("fe", "css")

	if len(s.Categories["fe"]) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(s.Categories["fe"]))
	}

	// Adding duplicate is idempotent
	s.AddSkillToCategory("fe", "react")
	if len(s.Categories["fe"]) != 2 {
		t.Fatalf("expected still 2 skills after duplicate add, got %d", len(s.Categories["fe"]))
	}

	s.RemoveSkillFromCategory("fe", "react")
	if len(s.Categories["fe"]) != 1 {
		t.Fatalf("expected 1 skill after remove, got %d", len(s.Categories["fe"]))
	}
	if s.Categories["fe"][0] != "css" {
		t.Fatalf("expected remaining skill 'css', got %s", s.Categories["fe"][0])
	}
}

func TestSetGetNote(t *testing.T) {
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", t.TempDir())
	defer os.Setenv("HOME", oldHome)

	s, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	s.SetNote("my-skill", "a useful note")
	if note := s.GetNote("my-skill"); note != "a useful note" {
		t.Fatalf("expected 'a useful note', got %q", note)
	}
	// Empty string removes note
	s.SetNote("my-skill", "")
	if note := s.GetNote("my-skill"); note != "" {
		t.Fatalf("expected empty note after clearing, got %q", note)
	}
	if _, ok := s.Notes["my-skill"]; ok {
		t.Fatal("expected note key to be deleted after clearing")
	}
}

func TestSortedCategoryNames(t *testing.T) {
	s := &Store{Categories: map[string][]string{
		"z": {},
		"a": {},
		"m": {},
	}}
	names := s.SortedCategoryNames()
	if len(names) != 3 {
		t.Fatalf("expected 3 names, got %d", len(names))
	}
	if names[0] != "a" || names[1] != "m" || names[2] != "z" {
		t.Fatalf("expected sorted [a m z], got %v", names)
	}
}

func TestDeleteCategory(t *testing.T) {
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", t.TempDir())
	defer os.Setenv("HOME", oldHome)

	s, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	s.CreateCategory("tmp")
	s.DeleteCategory("tmp")
	if _, ok := s.Categories["tmp"]; ok {
		t.Fatal("expected category to be deleted")
	}
}