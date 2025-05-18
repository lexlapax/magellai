// ABOUTME: Domain types for search functionality including SearchResult and SearchMatch
// ABOUTME: Core business entities for session search operations and results

package domain

// SearchResult represents the result of a session search operation.
type SearchResult struct {
	Session *SessionInfo  `json:"session"`
	Matches []SearchMatch `json:"matches"`
}

// SearchMatch represents a single match within a search result.
type SearchMatch struct {
	Type     string `json:"type"`     // "message", "system_prompt", "name", "tag"
	Role     string `json:"role"`     // for messages: "user", "assistant", "system"
	Content  string `json:"content"`  // the actual matched content snippet
	Context  string `json:"context"`  // surrounding context
	Position int    `json:"position"` // message index if applicable
}

// SearchMatchType constants define the types of searchable content.
const (
	SearchMatchTypeMessage      = "message"
	SearchMatchTypeSystemPrompt = "system_prompt"
	SearchMatchTypeName         = "name"
	SearchMatchTypeTag          = "tag"
)

// NewSearchResult creates a new search result for a session.
func NewSearchResult(session *SessionInfo) *SearchResult {
	return &SearchResult{
		Session: session,
		Matches: []SearchMatch{},
	}
}

// AddMatch adds a search match to the result.
func (sr *SearchResult) AddMatch(match SearchMatch) {
	sr.Matches = append(sr.Matches, match)
}

// GetMatchCount returns the total number of matches.
func (sr *SearchResult) GetMatchCount() int {
	return len(sr.Matches)
}

// HasMatches returns true if there are any matches.
func (sr *SearchResult) HasMatches() bool {
	return len(sr.Matches) > 0
}

// GetMatchesByType returns all matches of a specific type.
func (sr *SearchResult) GetMatchesByType(matchType string) []SearchMatch {
	matches := []SearchMatch{}
	for _, match := range sr.Matches {
		if match.Type == matchType {
			matches = append(matches, match)
		}
	}
	return matches
}

// NewSearchMatch creates a new search match.
func NewSearchMatch(matchType, role, content, context string, position int) SearchMatch {
	return SearchMatch{
		Type:     matchType,
		Role:     role,
		Content:  content,
		Context:  context,
		Position: position,
	}
}