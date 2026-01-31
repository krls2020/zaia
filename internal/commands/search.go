package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zeropsio/zaia/internal/knowledge"
	"github.com/zeropsio/zaia/internal/output"
)

// NewSearch creates the search command for BM25 knowledge search.
func NewSearch() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search [query...]",
		Short: "Search knowledge base",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			getURI, _ := cmd.Flags().GetString("get")

			// --get mode: direct document lookup by URI
			if getURI != "" {
				store := knowledge.GetEmbeddedStore()
				doc, err := store.Get(getURI)
				if err != nil {
					return output.Err("NOT_FOUND", "Document not found", "", nil)
				}
				return output.Sync(map[string]interface{}{
					"uri":     doc.URI,
					"title":   doc.Title,
					"content": doc.Content,
				})
			}

			if len(args) == 0 {
				return fmt.Errorf("requires at least 1 arg(s), only received 0")
			}

			query := strings.Join(args, " ")
			limit, _ := cmd.Flags().GetInt("limit")

			store := knowledge.GetEmbeddedStore()
			results := store.Search(query, limit)
			suggestions := store.GenerateSuggestions(query, results)

			expandedQuery := knowledge.ExpandQuery(query)

			// Build results list
			resultList := make([]interface{}, len(results))
			for i, r := range results {
				resultList[i] = map[string]interface{}{
					"uri":     r.URI,
					"title":   r.Title,
					"score":   r.Score,
					"snippet": r.Snippet,
				}
			}

			// Build topResult (full content of #1 if score >= 1.0)
			var topResult interface{}
			if len(results) > 0 && results[0].Score >= 1.0 {
				doc, err := store.Get(results[0].URI)
				if err == nil {
					topResult = map[string]interface{}{
						"uri":     doc.URI,
						"title":   doc.Title,
						"content": doc.Content,
					}
				}
			}

			return output.Sync(map[string]interface{}{
				"query":         query,
				"expandedQuery": expandedQuery,
				"results":       resultList,
				"topResult":     topResult,
				"suggestions":   suggestions,
			})
		},
	}

	cmd.Flags().String("get", "", "Get document by URI")
	cmd.Flags().Int("limit", 5, "Max results (1-20)")
	return cmd
}
