package index

import (
	"strings"

	"github.com/rzzdr/marrow/internal/model"
)

type Conflict struct {
	NewLearning      model.Learning
	ConflictsWith    string // description of what it conflicts with
	ConflictingEntry string // ID or summary of the conflicting entry
}

func DetectConflicts(
	newLearning model.Learning,
	learnings model.LearningsFile,
	graveyard model.GraveyardFile,
) []Conflict {
	var conflicts []Conflict

	newTags := toSet(newLearning.Tags)
	newWords := extractWords(newLearning.Text)

	for _, g := range graveyard.Entries {
		gTags := toSet(g.Tags)
		gWords := extractWords(g.Approach + " " + g.Reason)

		if hasOverlap(newTags, gTags) || hasWordOverlap(newWords, gWords) {
			conflicts = append(conflicts, Conflict{
				NewLearning:      newLearning,
				ConflictsWith:    "graveyard",
				ConflictingEntry: g.ID + ": " + g.Approach,
			})
		}
	}

	if newLearning.Type == model.LearningProven {
		for _, l := range learnings.Assumptions {
			if hasTagOverlap(newLearning, l) || hasTextOverlap(newLearning, l) {
				conflicts = append(conflicts, Conflict{
					NewLearning:      newLearning,
					ConflictsWith:    "assumption",
					ConflictingEntry: l.ID + ": " + l.Text,
				})
			}
		}
	}

	if newLearning.Type == model.LearningAssumption {
		for _, l := range learnings.Proven {
			if hasTagOverlap(newLearning, l) || hasTextOverlap(newLearning, l) {
				conflicts = append(conflicts, Conflict{
					NewLearning:      newLearning,
					ConflictsWith:    "proven learning",
					ConflictingEntry: l.ID + ": " + l.Text,
				})
			}
		}
	}

	return conflicts
}

func toSet(items []string) map[string]bool {
	s := make(map[string]bool, len(items))
	for _, item := range items {
		s[strings.ToLower(item)] = true
	}
	return s
}

func extractWords(text string) map[string]bool {
	words := make(map[string]bool)
	for _, w := range strings.Fields(strings.ToLower(text)) {
		if len(w) > 3 {
			words[w] = true
		}
	}
	return words
}

func hasOverlap(a, b map[string]bool) bool {
	for k := range a {
		if b[k] {
			return true
		}
	}
	return false
}

func hasWordOverlap(a, b map[string]bool) bool {
	count := 0
	for k := range a {
		if b[k] {
			count++
		}
	}
	return count >= 2
}

func hasTagOverlap(a, b model.Learning) bool {
	aSet := toSet(a.Tags)
	bSet := toSet(b.Tags)
	return hasOverlap(aSet, bSet)
}

func hasTextOverlap(a, b model.Learning) bool {
	aWords := extractWords(a.Text)
	bWords := extractWords(b.Text)
	return hasWordOverlap(aWords, bWords)
}
