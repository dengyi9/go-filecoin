package binpack

import (
	"context"
	"gx/ipfs/QmVmDhyTTUcQXFD1rRQ64fGLMSAoaQvNH3hwuaCFAPq2hy/errors"
)

// Bin-packing problem: https://en.wikipedia.org/wiki/Bin_packing_problem

// ErrItemTooLarge signals that an item was larger than the bin size so will never fit any bin.
var ErrItemTooLarge = errors.New("item too large for bin")

// Bin is a container into which Items are packed.
type Bin interface{}

// Item is implemented by types which are packed into Bins.
type Item interface{}

// Space is the size unit.
type Space uint

// NaivePacker implements a single-bin packing strategy.
type NaivePacker struct {
	bin    Bin
	binner Binner
}

// Future work to include implementing FirstFitPacker, then ModifiedFirstFitPacker, as needed.

// Packer is implemented by types defining a packing strategy.
type Packer interface {
	Init(Binner) (Bin, error)
	AddItem(context.Context, Item) (Bin, error)
}

// Binner is implemented by types which handle concrete binning of items.
type Binner interface {
	AddItem(context.Context, Item, Bin) error
	BinSize() Space
	CloseBin(Bin)
	ItemSize(Item) Space
	NewBin() (Bin, error)
	SpaceAvailable(bin Bin) Space
}

// Init implements Packer, associating it with a concrete Binner.
func (np *NaivePacker) Init(binner Binner) (Bin, error) {
	np.binner = binner
	bin, err := binner.NewBin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new bin")
	}
	np.bin = bin
	return bin, nil
}

// NewNaivePacker allocates and initializes a NaivePacker and an initial Binner, returning them along with any error.
func NewNaivePacker(binner Binner) (Packer, Bin, error) {
	packer := &NaivePacker{}
	bin, err := packer.Init(binner)
	return packer, bin, errors.Wrap(err, "failed to initialize packer")
}

// AddItem takes a context and an Item, and adds the Item according to the naive packing strategy.
// Returns the Bin to which item was added, and any error.
func (np *NaivePacker) AddItem(ctx context.Context, item Item) (Bin, error) {
	binner := np.binner
	bin := np.bin
	size := binner.ItemSize(item)

	if size > binner.BinSize() {
		return nil, ErrItemTooLarge
	}
	if size > binner.SpaceAvailable(bin) {
		binner.CloseBin(bin)
		newBin, err := binner.NewBin()
		if err != nil {
			return nil, errors.Wrap(err, "failed to create new bin")
		}
		bin = newBin
	}

	return bin, binner.AddItem(ctx, item, bin)
}