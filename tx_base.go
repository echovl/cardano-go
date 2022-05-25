package cardano

type Hash28 [28]byte

type Hash32 [32]byte

type AddressBytes []byte

type AddrKeyHash Hash28

type PoolKeyHash Hash28

type Coin uint64

type UnitInterval interface{}

type Uint64 *uint64

func NewUint64(u uint64) Uint64 {
	return Uint64(&u)
}
