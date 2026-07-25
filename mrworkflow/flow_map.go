package mrworkflow

import (
	"maps"
	"slices"
)

type (
	// FlowMap - интерфейс управления переходами между статусами.
	// Определяет, какие статусы зарегистрированы и какие переходы между ними допустимы.
	//
	// Методы, возвращающие списки, отдают копию: изменение выданного среза
	// на состояние карты не влияет. Пустой список отдаётся как nil, а не как пустой
	// срез, поэтому по ответу PossibleToStatuses и PossibleFromStatuses нельзя
	// отличить статус без переходов от незарегистрированного - для этого есть Exists.
	FlowMap[Status ~uint8] interface {
		Registered() []Status
		Exists(status Status) bool
		IsPossible(from, to Status) bool
		PossibleToStatuses(from Status) []Status
		PossibleFromStatuses(to Status) []Status
	}

	// FlowNode - описывает допустимые переходы из одного статуса в другие.
	// From - исходный статус.
	// To - список статусов, в которые разрешён переход из From.
	FlowNode[Status ~uint8] struct {
		From Status
		To   []Status
	}

	statusFlow[Status ~uint8] struct {
		fromToMap     map[Status][]Status
		toFromMap     map[Status][]Status
		registeredMap map[Status]bool
		registered    []Status
	}
)

// NewFlowMap - создаёт карту допустимых переходов между статусами.
// Параметр list - список узлов переходов (FlowNode), определяющих граф состояний.
// Автоматически строит двунаправленную карту: from→to и to→from.
//
// Каждый исходный статус описывается одним узлом: список - это описание графа целиком,
// а не набор дополнений к нему. Повтор From отбрасывается целиком, поэтому выигрывает
// первое вхождение: переходы повтора не попадают ни в прямую карту, ни в обратную,
// и статус, который называл только повтор, зарегистрированным не считается.
//
// Отбрасывается именно узел целиком, а не одна из карт: если оставить повтору обратную
// карту, то прямая сохранила бы переходы первого узла, а обратная - обоих,
// и IsPossible разошёлся бы с PossibleFromStatuses.
func NewFlowMap[Status ~uint8](list []FlowNode[Status]) FlowMap[Status] {
	fromToMap := make(map[Status][]Status, len(list))
	toFromMap := make(map[Status][]Status, len(list))
	registeredMap := make(map[Status]bool, len(list))

	for _, item := range list {
		// повтор узла отбрасывается до заполнения карт: проверяется наличие ключа,
		// а не registeredMap, т.к. в последнем статус мог быть отмечен как цель
		// перехода, а собственного узла ещё не иметь. Значение ключа при этом
		// может быть nil (узел без переходов), поэтому годится только запрос с ok
		if _, ok := fromToMap[item.From]; ok {
			continue
		}

		// список переходов копируется, а не берётся как есть: иначе состояние карты
		// осталось бы связанным с FlowNode.To вызывающего, и его изменение после
		// создания карты меняло бы набор допустимых переходов.
		// Пустой список сводится к nil отдельной проверкой, а не одним slices.Clone:
		// Clone сохраняет nil только для nil, а срез нулевой длины ([]Status{}) вернул бы
		// таким же не-nil, и узел без переходов ответил бы пустым срезом вместо nil
		// в нарушение общего правила (см. FlowMap)
		var toStatuses []Status

		if len(item.To) > 0 {
			toStatuses = slices.Clone(item.To)
		}

		fromToMap[item.From] = toStatuses
		registeredMap[item.From] = true

		for _, to := range item.To {
			toFromMap[to] = append(toFromMap[to], item.From)
			registeredMap[to] = true
		}
	}

	registered := slices.Collect(maps.Keys(registeredMap))

	return &statusFlow[Status]{
		fromToMap:     fromToMap,
		toFromMap:     toFromMap,
		registeredMap: registeredMap,
		registered:    registered,
	}
}

// Registered - возвращает список зарегистрированных статусов в карте.
func (f *statusFlow[Status]) Registered() []Status {
	return slices.Clone(f.registered)
}

// Exists - сообщает, имеется ли данный статус в карте статусов.
func (f *statusFlow[Status]) Exists(status Status) bool {
	return f.registeredMap[status]
}

// IsPossible - сообщает, возможно ли переключить данный статус в указанный статус.
func (f *statusFlow[Status]) IsPossible(from, to Status) bool {
	toStatuses, ok := f.fromToMap[from]
	if !ok {
		return false
	}

	for i := range toStatuses {
		if toStatuses[i] == to {
			return true
		}
	}

	return false
}

// PossibleToStatuses - возвращает список статусов в которые можно переключить указанный статус.
func (f *statusFlow[Status]) PossibleToStatuses(from Status) []Status {
	return slices.Clone(f.fromToMap[from])
}

// PossibleFromStatuses - возвращает список статусов из которых можно переключиться в указанный статус.
func (f *statusFlow[Status]) PossibleFromStatuses(to Status) []Status {
	return slices.Clone(f.toFromMap[to])
}
