package skeleton

// Option configures public Skel generation.
type Option struct {
	// PubOnly limits output to declarations in the public contract.
	PubOnly bool
	// Out is the output directory for generated Skel files.
	Out string
}
