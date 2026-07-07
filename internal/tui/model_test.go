package tui

import (
	"testing"

	"skill-management/internal/repo"
)

func TestFilteredSkills_ByCategory(t *testing.T) {
	skills := []repo.Skill{
		{Name: "react", Category: "frontend"},
		{Name: "css", Category: "frontend"},
		{Name: "api", Category: "backend"},
	}
	m := model{
		skills:     skills,
		panelFocus: 0,
		catCursor:  0,
		categories: []repo.Category{
			{Name: "frontend", Skills: skills[:2]},
			{Name: "backend", Skills: skills[2:]},
		},
	}

	result := m.filteredSkills()
	if len(result) != 2 {
		t.Fatalf("expected 2 skills in frontend filter, got %d", len(result))
	}
	if result[0].Name != "react" {
		t.Fatalf("expected first skill 'react', got %s", result[0].Name)
	}
	if result[1].Name != "css" {
		t.Fatalf("expected second skill 'css', got %s", result[1].Name)
	}
}

func TestFilteredSkills_AllSkills(t *testing.T) {
	skills := []repo.Skill{
		{Name: "react", Category: "frontend"},
		{Name: "api", Category: "backend"},
	}
	m := model{
		skills:     skills,
		panelFocus: 1, // skill panel focused = no filter
		categories: []repo.Category{
			{Name: "frontend", Skills: skills[:1]},
			{Name: "backend", Skills: skills[1:]},
		},
	}

	result := m.filteredSkills()
	if len(result) != 2 {
		t.Fatalf("expected all 2 skills with panelFocus=1, got %d", len(result))
	}
}

func TestFilteredSkills_EmptyCategory(t *testing.T) {
	skills := []repo.Skill{
		{Name: "react", Category: "frontend"},
	}
	m := model{
		skills:     skills,
		panelFocus: 0,
		catCursor:  1, // second category (backend) is empty or doesn't exist
		categories: []repo.Category{
			{Name: "frontend", Skills: skills},
		},
	}

	// catCursor=1 is out of range, so filteredSkills should return all skills
	result := m.filteredSkills()
	if len(result) != 1 {
		t.Fatalf("expected 1 skill (fallback to all), got %d", len(result))
	}
}

func TestFilteredSkills_NoCategories(t *testing.T) {
	skills := []repo.Skill{
		{Name: "react", Category: "frontend"},
	}
	m := model{
		skills:     skills,
		panelFocus: 0,
		catCursor:  0,
		categories: nil,
	}

	result := m.filteredSkills()
	if len(result) != 1 {
		t.Fatalf("expected 1 skill with no categories, got %d", len(result))
	}
}

func TestFilteredSkills_EmptySkills(t *testing.T) {
	m := model{
		skills:     nil,
		panelFocus: 0,
		catCursor:  0,
		categories: []repo.Category{
			{Name: "frontend", Skills: nil},
		},
	}

	result := m.filteredSkills()
	if len(result) != 0 {
		t.Fatalf("expected 0 skills, got %d", len(result))
	}
}