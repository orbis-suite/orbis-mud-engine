package conditions

import (
	"regexp"
	"strings"

	"example.com/mud/world/entities"
)

type MessageContains struct {
	MessageRegex string
}

var _ entities.Condition = &MessageContains{}

func (mc *MessageContains) Id() entities.ConditionType {
	return entities.ConditionMessageMatches
}

func (mc *MessageContains) Check(ev *entities.Event) (bool, error) {
	if ev == nil || ev.Message == "" {
		return false, nil
	}

	re, err := regexp.Compile(mc.MessageRegex)
	if err != nil {
		return false, err
	}

	return re.MatchString(strings.ToLower(ev.Message)), nil
}
