package mrenum

import (
    "database/sql/driver"
    "encoding/json"

    "github.com/mondegor/go-webcore/mrcore"
)

const (
    _ ItemStatus = iota
    ItemStatusDraft
    ItemStatusEnabled
    ItemStatusDisabled
    ItemStatusRemoved

    enumNameItemStatus = "ItemStatus"
)

type (
    ItemStatus uint8
)

var (
    itemStatusName = map[ItemStatus]string{
        ItemStatusDraft: "DRAFT",
        ItemStatusEnabled: "ENABLED",
        ItemStatusDisabled: "DISABLED",
        ItemStatusRemoved: "REMOVED",
    }

    itemStatusValue = map[string]ItemStatus{
        "DRAFT": ItemStatusDraft,
        "ENABLED": ItemStatusEnabled,
        "DISABLED": ItemStatusDisabled,
        "REMOVED": ItemStatusRemoved,
    }

    ItemStatusFlow = StatusFlow{
        ItemStatusDraft: {
            ItemStatusEnabled,
            ItemStatusDisabled,
            ItemStatusRemoved,
        },
        ItemStatusEnabled: {
            ItemStatusDisabled,
            ItemStatusRemoved,
        },
        ItemStatusDisabled: {
            ItemStatusEnabled,
            ItemStatusRemoved,
        },
        ItemStatusRemoved: {},
    }
)

func (e *ItemStatus) ParseAndSet(value string) error {
    if parsedValue, ok := itemStatusValue[value]; ok {
        *e = parsedValue
        return nil
    }

    return mrcore.FactoryErrInternalMapValueNotFound.New(value, enumNameItemStatus)
}

func (e ItemStatus) String() string {
    return itemStatusName[e]
}

func (e ItemStatus) MarshalJSON() ([]byte, error) {
    return json.Marshal(e.String())
}

func (e *ItemStatus) UnmarshalJSON(data []byte) error {
    var value string

    if err := json.Unmarshal(data, &value); err != nil {
        return err
    }

    return e.ParseAndSet(value)
}

// Scan implements the Scanner interface.
func (e *ItemStatus) Scan(value any) error {
    if val, ok := value.(string); ok {
        return e.ParseAndSet(val)
    }

    return mrcore.FactoryErrInternalTypeAssertion.New(enumNameItemStatus, value)
}

func (e ItemStatus) Value() (driver.Value, error) {
    return e.String(), nil
}

func ParseItemStatusList(items []string) ([]ItemStatus, error) {
    var tmp ItemStatus
    parsedItems := make([]ItemStatus, len(items))

    for i := range items {
        if err := tmp.ParseAndSet(items[i]); err != nil {
            return nil, err
        }

        parsedItems[i] = tmp
    }

    return parsedItems, nil
}
