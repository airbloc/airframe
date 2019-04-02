package database

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestQuery_Equals(t *testing.T) {
	q, err := queryFromJson(`{"name": {"eq": "Hyojun Kim"}}`)
	require.NoError(t, err)
	require.Equal(t, 1, len(q.Conditions))
	require.Equal(t, OperatorType(OpEquals), q.Conditions[0].Type)
	require.Equal(t, "name", q.Conditions[0].Field)
	require.Equal(t, "Hyojun Kim", q.Conditions[0].Operand)
}

func TestQuery_EqualsAbbr(t *testing.T) {
	q, err := queryFromJson(`{"name": "Hyojun Kim"}`)
	require.NoError(t, err)
	require.Equal(t, 1, len(q.Conditions))
	require.Equal(t, OperatorType(OpEquals), q.Conditions[0].Type)
	require.Equal(t, "name", q.Conditions[0].Field)
	require.Equal(t, "Hyojun Kim", q.Conditions[0].Operand)
}

func TestQuery_Gte(t *testing.T) {
	q, err := queryFromJson(`{"age": {"gte": 20}}`)
	require.NoError(t, err)
	require.Equal(t, 1, len(q.Conditions))
	require.Equal(t, OperatorType(OpGreaterThanOrEqual), q.Conditions[0].Type)
	require.Equal(t, "age", q.Conditions[0].Field)
	require.Equal(t, 20.0, q.Conditions[0].Operand)
}
