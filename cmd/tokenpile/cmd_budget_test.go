package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cdimonaco/tokenpile/internal/store"
)

func TestIntegration_Budget_Set(t *testing.T) {
	s := newTestStore(t)

	out, err := runBudgetCmd(t, s, "budget", "set", "--issue", "1", "--repo", "owner/repo", "--amount", "10.00")
	require.NoError(t, err)
	assert.Contains(t, out, "$10.00")

	budget, err := s.GetBudget(context.Background(), "owner/repo", 1)
	require.NoError(t, err)
	require.NotNil(t, budget)
	assert.InEpsilon(t, 10.00, *budget, 0.001)
}

func TestIntegration_Budget_ZeroAmount_Fails(t *testing.T) {
	s := newTestStore(t)

	_, err := runBudgetCmd(t, s, "budget", "set", "--issue", "1", "--repo", "owner/repo", "--amount", "0")
	assert.Error(t, err)
}

func TestIntegration_Budget_NegativeAmount_Fails(t *testing.T) {
	s := newTestStore(t)

	_, err := runBudgetCmd(t, s, "budget", "set", "--issue", "1", "--repo", "owner/repo", "--amount", "-5")
	assert.Error(t, err)
}

func TestIntegration_Budget_Unset(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.SetBudget(ctx, "owner/repo", 1, 20.00))

	_, err := runBudgetCmd(t, s, "budget", "unset", "--issue", "1", "--repo", "owner/repo")
	require.NoError(t, err)

	_, err = s.GetBudget(ctx, "owner/repo", 1)
	require.ErrorIs(t, err, store.ErrBudgetNotFound)
}

func TestIntegration_Budget_Overwrite(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.SetBudget(ctx, "owner/repo", 1, 10.00))

	_, err := runBudgetCmd(t, s, "budget", "set", "--issue", "1", "--repo", "owner/repo", "--amount", "25.00")
	require.NoError(t, err)

	budget, err := s.GetBudget(ctx, "owner/repo", 1)
	require.NoError(t, err)
	require.NotNil(t, budget)
	assert.InEpsilon(t, 25.00, *budget, 0.001)
}
