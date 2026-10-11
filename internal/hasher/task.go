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
		skelName := s.domain.ReferenceName(audience.Actor)
		name := skelName[strings.LastIndexByte(skelName, '.')+1:]
		hash := ""
		if actor := s.actorBySkel[skelName]; actor != nil {
			hash = s.actorHash(actor)
		}
		values = append(values, &_ActorRefHashValue{
			Name:     name,
			SkelName: skelName,
			Hash:     hash,
		})
	}
	return values
}
