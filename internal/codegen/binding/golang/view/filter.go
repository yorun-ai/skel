package view

import "go.yorun.ai/skel/schema"

func filterNonPubData(dataList []*schema.Data) []*schema.Data {
	filtered := make([]*schema.Data, 0, len(dataList))
	for _, data := range dataList {
		if !data.Pub {
			filtered = append(filtered, data)
		}
	}
	return filtered
}

func filterNonPubActors(actors []*schema.Actor) []*schema.Actor {
	filtered := make([]*schema.Actor, 0, len(actors))
	for _, actor := range actors {
		if !actor.Pub {
			filtered = append(filtered, actor)
		}
	}
	return filtered
}

func filterNonPubResources(resources []*schema.Resource) []*schema.Resource {
	filtered := make([]*schema.Resource, 0, len(resources))
	for _, resource := range resources {
		if !resource.Pub {
			filtered = append(filtered, resource)
		}
	}
	return filtered
}
