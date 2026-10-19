package errors_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mrerrors "github.com/mondegor/go-core/mrtype/errors"
)

func TestParamIncorrectError_Error(t *testing.T) {
	t.Parallel()

	err := mrerrors.NewParamIncorrectError("Int64", errors.New("some cause"))

	assert.EqualError(t, err, "param contains incorrect value (type=Int64): some cause")
}

func TestParamIncorrectError_UnwrapSentinel(t *testing.T) {
	t.Parallel()

	cause := errors.New("some cause")
	err := mrerrors.NewParamIncorrectError("Int64", cause)

	require.ErrorIs(t, err, cause)
}

func TestParamIncorrectError_UnwrapNumError(t *testing.T) {
	t.Parallel()

	_, cause := strconv.ParseInt("abc", 10, 64)
	err := mrerrors.NewParamIncorrectError("Int64", cause)

	var numErr *strconv.NumError

	require.ErrorAs(t, err, &numErr)
	require.ErrorIs(t, err, strconv.ErrSyntax)
	assert.Equal(t, "abc", numErr.Num)
	assert.EqualError(t, err, `param contains incorrect value (type=Int64): strconv.ParseInt: parsing "abc": invalid syntax`)
}

func TestParamIncorrectError_UnwrapNested(t *testing.T) {
	t.Parallel()

	err := mrerrors.NewParamIncorrectError("PageParams.Index", mrerrors.NewParamLenMaxError("Uint64", 20))

	var lenMaxErr *mrerrors.ParamLenMaxError

	require.ErrorAs(t, err, &lenMaxErr)
	assert.Equal(t, "Uint64", lenMaxErr.Type)
	assert.EqualError(
		t,
		err,
		"param contains incorrect value (type=PageParams.Index): "+
			"param has value length greater than max characters (type=Uint64, maxLen=20)",
	)
}
