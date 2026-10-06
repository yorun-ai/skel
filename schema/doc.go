// Package schema defines normalized Skel language contracts and provides queries,
// JSON encoding, validation, and compatibility comparisons. These contracts also
// describe skelc schema list, get, snapshot, and diff output. Command failures use
// the contract exposed by go.yorun.ai/skel/cmd/skelc/output.
//
// Schema documents describe normalized language contracts independently of
// compiler models and source loading. Use go.yorun.ai/skel/api.ProjectSchema
// to project a semantic model into this contract representation.
package schema
