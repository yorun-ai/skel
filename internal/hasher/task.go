package hasher

import (
	"fmt"
	"strings"

	"go.yorun.ai/skel/schema"
)

func (s *_HashState) triggerHash(trigger *schema.TaskTrigger) string {
	return s.hashValue(_TriggerHashValue{
		Name:               trigger.Name,
		SkelName:           trigger.SkelName,
		Description:        trigger.Description,
		Deprecated:         trigger.Deprecated,
		DeprecatedReason:   trigger.DeprecatedReason,
		InputDescription:   trigger.InputDescription,
		ArgumentsSensitive: trigger.ArgumentsSensitive,
		Arguments:          s.buildArgumentHashValues(trigger.Arguments),
	})
}

func (s *_HashState) taskHash(task *schema.Task) string {
	return s.memoHash("task", task.SkelName, func() string {
		for _, trigger := range task.Triggers {
			trigger.Hash = s.triggerHash(trigger)
		}
		return s.hashValue(_TaskHashValue{
			Name:             task.Name,
			SkelName:         task.SkelName,
			Description:      task.Description,
			Deprecated:       task.Deprecated,
			DeprecatedReason: task.DeprecatedReason,
			Triggers: buildNamedValues(task.Triggers,
				func(trigger *schema.TaskTrigger) string { return trigger.SkelName },
				func(trigger *schema.TaskTrigger) string { return trigger.Hash }),
		})
	})
}

func (s *_HashState) memoHash(kind string, skelName string, build func() string) string {
	key := fmt.Sprintf("%s:%s", kind, skelName)
	if hash, ok := s.cache[key]; ok {
		return hash
	}
	if s.status[key] {
		return s.hashValue(struct {
			Kind     string `json:"kind"`
			SkelName string `json:"skelName"`
		}{
			Kind:     kind,
			SkelName: skelName,
		})
	}
	s.status[key] = true
	hash := build()
	s.status[key] = false
	s.cache[key] = hash
	return hash
}

func (s *_HashState) hashValue(value any) string {
	if s.err != nil {
		return ""
	}
	hash, err := hashValue(value)
	if err != nil {
		s.err = err
		return ""
	}
	return hash
}

func (s *_HashState) buildActorAudienceHashValues(audiences []*schema.ActorAudience) []*_ActorRefHashValue {
	values := make([]*_ActorRefHashValue, 0, len(audiences))
	for _, audience := range audiences {
		name, skelName := s.actorRefNames(audience.Actor)
		hash := ""
		if actor := s.actorBySkel[skelName]; actor != nil {
			hash = s.actorHash(actor)
		}
		values = append(values, &_ActorRefHashValue{
			Name:     name,
			SkelName: skelName,
			Hash:     hash,
			Via:      audience.Via,
		})
	}
	return values
}

func (s *_HashState) actorRefNames(actorName string) (string, string) {
	if index := strings.LastIndexByte(actorName, '.'); index >= 0 {
		alias, baseName := actorName[:index], actorName[index+1:]
		for _, import_ := range s.domain.Imports() {
			if import_.Alias == alias {
				return baseName, import_.Name + "." + baseName
			}
		}
	}
	return actorName, s.domain.Name() + "." + actorName
}
