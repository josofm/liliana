package user

import (
	"fmt"
	entity "github.com/josofm/liliana/internal/entity/user"
	"github.com/stretchr/testify/require"
	"testing"
)

func testSearch(t *testing.T, r Repository) {
	t.Helper()
	for _, u := range []*entity.User{
		{Name: "Marina", Email: "sea@example.com", Password: "secret"},
		{Name: "Rafael", Email: "MARINA@example.com", Password: "secret"},
		{Name: "Literal", Email: "100%_literal@example.com", Password: "secret"},
	} {
		require.NoError(t, r.Create(u))
	}
	results, err := r.Search("mArInA")
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "Rafael", results[0].Name)
	for _, u := range results {
		require.Empty(t, u.Password)
	}
	results, err = r.Search("Rafael")
	require.NoError(t, err)
	require.Empty(t, results)
	results, err = r.Search("%_")
	require.NoError(t, err)
	require.Len(t, results, 1)
	results, err = r.Search("nobody-found")
	require.NoError(t, err)
	require.NotNil(t, results)
	require.Empty(t, results)
	for n := 0; n < 25; n++ {
		require.NoError(t, r.Create(&entity.User{Name: fmt.Sprintf("Search %02d", n), Email: fmt.Sprintf("search%d@example.com", n), Password: "secret"}))
	}
	results, err = r.Search("search")
	require.NoError(t, err)
	require.Len(t, results, 20)
	require.Equal(t, "Search 00", results[0].Name)
	require.Equal(t, "Search 19", results[19].Name)
}
func TestMemorySearch(t *testing.T) { testSearch(t, NewInMemoryRepo()) }
