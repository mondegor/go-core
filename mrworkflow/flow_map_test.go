package mrworkflow_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mondegor/go-core/mrworkflow"
)

type (
	testStatus uint8
)

const (
	testStatusDraft testStatus = iota + 1
	testStatusPublished
	testStatusArchived
	testStatusUnregistered
)

// testFlowMap - карта переходов draft -> published -> archived.
func testFlowMap(t *testing.T) mrworkflow.FlowMap[testStatus] {
	t.Helper()

	return mrworkflow.NewFlowMap(
		[]mrworkflow.FlowNode[testStatus]{
			{From: testStatusDraft, To: []testStatus{testStatusPublished}},
			{From: testStatusPublished, To: []testStatus{testStatusArchived}},
		},
	)
}

// TestStatusFlow_Registered_ReturnsCopy - выданный срез принадлежит вызывающему:
// его изменение на состав зарегистрированных статусов карты не влияет.
//
// Порядок статусов не проверяется: список собирается обходом карты и потому
// от запуска к запуску различается (см. NewFlowMap).
func TestStatusFlow_Registered_ReturnsCopy(t *testing.T) {
	t.Parallel()

	flowMap := testFlowMap(t)
	want := []testStatus{testStatusDraft, testStatusPublished, testStatusArchived}

	got := flowMap.Registered()
	got[0] = 0

	assert.ElementsMatch(t, want, flowMap.Registered())
}

// TestStatusFlow_PossibleToStatuses - проверяет, что список переходов из статуса отдаётся,
// а незарегистрированный статус отвечает nil.
func TestStatusFlow_PossibleToStatuses(t *testing.T) {
	t.Parallel()

	flowMap := testFlowMap(t)

	assert.Equal(t, []testStatus{testStatusPublished}, flowMap.PossibleToStatuses(testStatusDraft))
	assert.Nil(t, flowMap.PossibleToStatuses(testStatusUnregistered))
	// конечный статус: переходов из него нет, поэтому ответ такой же, как у незарегистрированного
	assert.Nil(t, flowMap.PossibleToStatuses(testStatusArchived))
}

// TestStatusFlow_PossibleFromStatuses - проверяет, что список статусов, из которых
// можно переключиться в указанный, отдаётся, а незарегистрированный статус отвечает nil.
func TestStatusFlow_PossibleFromStatuses(t *testing.T) {
	t.Parallel()

	flowMap := testFlowMap(t)

	assert.Equal(t, []testStatus{testStatusDraft}, flowMap.PossibleFromStatuses(testStatusPublished))
	assert.Equal(t, []testStatus{testStatusPublished}, flowMap.PossibleFromStatuses(testStatusArchived))
	assert.Nil(t, flowMap.PossibleFromStatuses(testStatusUnregistered))
	// начальный статус: перейти в него неоткуда
	assert.Nil(t, flowMap.PossibleFromStatuses(testStatusDraft))
}

// TestStatusFlow_PossibleToStatuses_ReturnsCopy - выданный срез принадлежит вызывающему:
// его изменение на переходы карты не влияет.
func TestStatusFlow_PossibleToStatuses_ReturnsCopy(t *testing.T) {
	t.Parallel()

	flowMap := testFlowMap(t)

	got := flowMap.PossibleToStatuses(testStatusDraft)
	got[0] = testStatusArchived

	assert.Equal(t, []testStatus{testStatusPublished}, flowMap.PossibleToStatuses(testStatusDraft))
	assert.False(t, flowMap.IsPossible(testStatusDraft, testStatusArchived))
}

// TestStatusFlow_PossibleFromStatuses_ReturnsCopy - выданный срез принадлежит вызывающему:
// его изменение на переходы карты не влияет.
func TestStatusFlow_PossibleFromStatuses_ReturnsCopy(t *testing.T) {
	t.Parallel()

	flowMap := testFlowMap(t)

	got := flowMap.PossibleFromStatuses(testStatusPublished)
	got[0] = testStatusArchived

	assert.Equal(t, []testStatus{testStatusDraft}, flowMap.PossibleFromStatuses(testStatusPublished))
}

// TestStatusFlow_SourceListIsCopied - проверяет, что карта не связана со списком узлов
// вызывающего: изменение FlowNode.To после создания карты набор переходов не меняет.
func TestStatusFlow_SourceListIsCopied(t *testing.T) {
	t.Parallel()

	toStatuses := []testStatus{testStatusPublished}
	flowMap := mrworkflow.NewFlowMap(
		[]mrworkflow.FlowNode[testStatus]{
			{From: testStatusDraft, To: toStatuses},
		},
	)

	toStatuses[0] = testStatusArchived

	assert.Equal(t, []testStatus{testStatusPublished}, flowMap.PossibleToStatuses(testStatusDraft))
	assert.True(t, flowMap.IsPossible(testStatusDraft, testStatusPublished))
	assert.False(t, flowMap.IsPossible(testStatusDraft, testStatusArchived))
}

// TestStatusFlow_DuplicateNodeIgnored - проверяет, что повторный узел отбрасывается
// целиком: выигрывает первое вхождение From, а переходы повтора не попадают
// ни в прямую карту, ни в обратную, поэтому они остаются согласованными.
func TestStatusFlow_DuplicateNodeIgnored(t *testing.T) {
	t.Parallel()

	flowMap := mrworkflow.NewFlowMap(
		[]mrworkflow.FlowNode[testStatus]{
			{From: testStatusDraft, To: []testStatus{testStatusPublished}},
			{From: testStatusDraft, To: []testStatus{testStatusArchived}},
		},
	)

	assert.Equal(t, []testStatus{testStatusPublished}, flowMap.PossibleToStatuses(testStatusDraft))
	assert.True(t, flowMap.IsPossible(testStatusDraft, testStatusPublished))
	assert.False(t, flowMap.IsPossible(testStatusDraft, testStatusArchived))

	// обратная карта согласована с прямой: переход повтора не попал и в неё
	assert.Equal(t, []testStatus{testStatusDraft}, flowMap.PossibleFromStatuses(testStatusPublished))
	assert.Nil(t, flowMap.PossibleFromStatuses(testStatusArchived))

	// статус, которого называл только повтор, зарегистрированным не считается
	assert.False(t, flowMap.Exists(testStatusArchived))
	assert.ElementsMatch(t, []testStatus{testStatusDraft, testStatusPublished}, flowMap.Registered())
}

// TestStatusFlow_DuplicateNodeIgnored_TargetRegisteredEarlier - проверяет, что повтор
// определяется по собственному узлу статуса, а не по его регистрации: статус, отмеченный
// ранее как цель перехода, своим узлом описывается как обычно.
func TestStatusFlow_DuplicateNodeIgnored_TargetRegisteredEarlier(t *testing.T) {
	t.Parallel()

	flowMap := mrworkflow.NewFlowMap(
		[]mrworkflow.FlowNode[testStatus]{
			{From: testStatusDraft, To: []testStatus{testStatusPublished}},
			{From: testStatusPublished, To: []testStatus{testStatusArchived}},
		},
	)

	assert.Equal(t, []testStatus{testStatusArchived}, flowMap.PossibleToStatuses(testStatusPublished))
	assert.True(t, flowMap.IsPossible(testStatusPublished, testStatusArchived))
}

// TestStatusFlow_EmptyLists - проверяет, что пустые списки отдаются как nil,
// а не как пустой срез: у карты без узлов и у узла без переходов ответ одинаков.
func TestStatusFlow_EmptyLists(t *testing.T) {
	t.Parallel()

	t.Run("flow map without nodes", func(t *testing.T) {
		t.Parallel()

		flowMap := mrworkflow.NewFlowMap[testStatus](nil)

		assert.Nil(t, flowMap.Registered())
		assert.Nil(t, flowMap.PossibleToStatuses(testStatusDraft))
		assert.Nil(t, flowMap.PossibleFromStatuses(testStatusDraft))
	})

	t.Run("node without transitions", func(t *testing.T) {
		t.Parallel()

		flowMap := mrworkflow.NewFlowMap(
			[]mrworkflow.FlowNode[testStatus]{
				{From: testStatusDraft, To: []testStatus{}},
			},
		)

		assert.True(t, flowMap.Exists(testStatusDraft))
		assert.Nil(t, flowMap.PossibleToStatuses(testStatusDraft))
	})
}
