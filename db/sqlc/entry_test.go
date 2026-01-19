package db

import (
	"context"
	"tkbank/util"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createRandomEntry(t *testing.T, accountID int64) Entry {
	arg := CreateEntryParams{
		AccountID: accountID,
		Amount:    util.RandomMoney(),
	}

	entry, err := testQueries.CreateEntry(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, entry)

	// Compare if account.ID is equal to entry.AccountID
	require.Equal(t, accountID, entry.AccountID)
	// Compare if arg.Amount is equal to entry.Amount
	require.Equal(t, arg.Amount, entry.Amount)
	require.NotZero(t, entry.ID)
	require.NotZero(t, entry.CreatedAt)

	return entry
}

func TestCreateEntry(t *testing.T) {
	account := createRandomAccount(t)
	createRandomEntry(t, account.ID)
}

func TestGetEntry(t *testing.T) {
	account := createRandomAccount(t)

	entry1 := createRandomEntry(t, account.ID)
	entry2, err := testQueries.GetEntry(context.Background(), entry1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, entry2)

	require.Equal(t, entry1.ID, entry2.ID)
	require.Equal(t, entry1.AccountID, entry2.AccountID)
	require.Equal(t, entry1.Amount, entry2.Amount)

	createdAt1 := util.ConvertTimestamptzToTime(entry1.CreatedAt)
	createdAt2 := util.ConvertTimestamptzToTime(entry2.CreatedAt)
	require.WithinDuration(t, createdAt1, createdAt2, time.Second)
}

func TestListEntries(t *testing.T) {
	account := createRandomAccount(t)

	for i := 0; i < 10; i++ {
		createRandomEntry(t, account.ID)
	}

	arg := ListEntriesParams{
		AccountID: account.ID,
		Limit:     5,
		Offset:    5,
	}

	entries, err := testQueries.ListEntries(context.Background(), arg)

	require.NoError(t, err)
	require.Len(t, entries, 5)

	for _, entry := range entries {
		require.NotEmpty(t, entry)
		require.Equal(t, arg.AccountID, entry.AccountID)
	}

	// ERROR CASE: invalid query
	wrongArg := ListEntriesParams{
		AccountID: 0,
		Limit:     -1,
		Offset:    -1,
	}
	_, err = testQueries.ListEntries(context.Background(), wrongArg)
	require.Error(t, err)

	// Simulate database query error
	testQueries.db.Exec(context.Background(), "DROP TABLE entries")
	_, err = testQueries.ListEntries(context.Background(), arg)
	require.Error(t, err)

	// Restore the database
	testQueries.db.Exec(context.Background(), `
		CREATE TABLE entries (
			id SERIAL PRIMARY KEY,
			account_id BIGINT NOT NULL,
			amount BIGINT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
		)
	`)
}
