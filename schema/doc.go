// Package schema exposes the stable JSON wire contract emitted by skelc schema
// commands. It is a public facade over the internal schema implementation and
// can be used by integrations that consume schema list, get, snapshot, or diff
// output. Command failures use the contract exposed by go.yorun.ai/skel/cmd/skelc/output.
//
// Schema documents describe normalized language contracts independently of
// compiler models and source loading. Use go.yorun.ai/skel/api.ProjectSchema
// to project a semantic model into this contract representation.
package schema
