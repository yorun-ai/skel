// Package schema exposes the stable JSON wire contract emitted by skelc schema
// commands. It is a public facade over the internal schema implementation and
// can be used by integrations that consume schema list, get, snapshot, or diff
// output. Command failures use the contract exposed by go.yorun.ai/skel/cli.
//
// Schema documents are normalized projections of the semantic model exposed by
// go.yorun.ai/skel/model. The two representations remain separate so compiler
// model changes cannot implicitly alter the versioned schema wire contract.
package schema
