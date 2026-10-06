package view

import "go.yorun.ai/skel/internal/model"

func filterNonPubData(dataList []*model.Data) []*model.Data {
	filtered := make([]*model.Data, 0, len(dataList))
	for _, data := range dataList {
		if !data.Pub {
			filtered = append(filtered, data)
		}
	}
	return filtered
}

func filterNonPubActors(actors []*model.Actor) []*model.Actor {
	filtered := make([]*model.Actor, 0, len(actors))
	for _, actor := range actors {
		if !actor.Pub {
			filtered = append(filtered, actor)
		}
	}
	return filtered
}

func filterNonPubResources(resources []*model.Resource) []*model.Resource {
	filtered := make([]*model.Resource, 0, len(resources))
	for _, resource := range resources {
		if !resource.Pub {
			filtered = append(filtered, resource)
		}
	}
	return filtered
}
