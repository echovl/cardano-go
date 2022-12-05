package prefixes

// As specified in [CIP-5](https://github.com/cardano-foundation/CIPs/tree/master/CIP5)
//
// copied from cardano-addresses core/lib/Cardano/Codec/Bech32/Prefixes.hs

type Bech32Prefix = string

const (
	// -- * Addresses

	Addr      Bech32Prefix = "addr"
	AddrTest  Bech32Prefix = "addr_test"
	Script    Bech32Prefix = "script"
	Stake     Bech32Prefix = "stake"
	StakeTest Bech32Prefix = "stake_test"

	// -- * Hashes

	AddrPublicKeyHash        Bech32Prefix = "addr_vkh"
	StakePublicKeyHash       Bech32Prefix = "stake_vkh"
	AddrSharedPublicKeyHash  Bech32Prefix = "addr_shared_vkh"
	StakeSharedPublicKeyHash Bech32Prefix = "stake_shared_vkh"

	// -- * Keys for 1852H
	AddrPublicKey          Bech32Prefix = "addr_vk"
	AddrPrivateKey         Bech32Prefix = "addr_sk"
	AddrXPub               Bech32Prefix = "addr_xvk"
	AddrXPrv               Bech32Prefix = "addr_xsk"
	AddrExtendedPublicKey               = AddrXPub
	AddrExtendedPrivateKey              = AddrXPrv

	AcctPublicKey          Bech32Prefix = "acct_vk"
	AcctPrivateKey         Bech32Prefix = "acct_sk"
	AcctXPub               Bech32Prefix = "acct_xvk"
	AcctXPrv               Bech32Prefix = "acct_xsk"
	AcctExtendedPublicKey               = AcctXPub
	AcctExtendedPrivateKey              = AcctXPrv

	RootPublicKey          Bech32Prefix = "root_vk"
	RootPrivateKey         Bech32Prefix = "root_sk"
	RootXPub               Bech32Prefix = "root_xvk"
	RootXPrv               Bech32Prefix = "root_xsk"
	RootExtendedPublicKey               = RootXPub
	RootExtendedPrivateKey              = RootXPrv

	StakePublicKey          Bech32Prefix = "stake_vk"
	StakePrivateKey         Bech32Prefix = "stake_sk"
	StakeXPub               Bech32Prefix = "stake_xvk"
	StakeXPrv               Bech32Prefix = "stake_xsk"
	StakeExtendedPublicKey               = StakeXPub
	StakeExtendedPrivateKey              = StakeXPrv

	// -- * Keys for 1854H

	AddrSharedPublicKey          Bech32Prefix = "addr_shared_vk"
	AddrSharedPrivateKey         Bech32Prefix = "addr_shared_sk"
	AddrSharedXPub               Bech32Prefix = "addr_shared_xvk"
	AddrSharedXPrv               Bech32Prefix = "addr_shared_xsk"
	AddrSharedExtendedPublicKey               = AddrSharedXPub
	AddrSharedExtendedPrivateKey              = AddrSharedXPrv

	AcctSharedPublicKey          Bech32Prefix = "acct_shared_vk"
	AcctSharedPrivateKey         Bech32Prefix = "acct_shared_sk"
	AcctSharedXPub               Bech32Prefix = "acct_shared_xvk"
	AcctSharedXPrv               Bech32Prefix = "acct_shared_xsk"
	AcctSharedExtendedPublicKey               = AcctSharedXPub
	AcctSharedExtendedPrivateKey              = AcctSharedXPrv

	RootSharedPublicKey          Bech32Prefix = "root_shared_vk"
	RootSharedPrivateKey         Bech32Prefix = "root_shared_sk"
	RootSharedXPub               Bech32Prefix = "root_shared_xvk"
	RootSharedXPrv               Bech32Prefix = "root_shared_xsk"
	RootSharedExtendedPublicKey               = RootSharedXPub
	RootSharedExtendedPrivateKey              = RootSharedXPrv

	StakeSharedPublicKey          Bech32Prefix = "stake_shared_vk"
	StakeSharedPrivateKey         Bech32Prefix = "stake_shared_sk"
	StakeSharedXPub               Bech32Prefix = "stake_shared_xvk"
	StakeSharedXPrv               Bech32Prefix = "stake_shared_xsk"
	StakeSharedExtendedPublicKey               = StakeSharedXPub
	StakeSharedExtendedPrivateKey              = StakeSharedXPrv

	// -- * Keys for 1855H
	PolicyPublicKey          Bech32Prefix = "policy_vk"
	PolicyPrivateKey         Bech32Prefix = "policy_sk"
	PolicyXPub               Bech32Prefix = "policy_xvk"
	PolicyXPrv               Bech32Prefix = "policy_xsk"
	PolicyExtendedPublicKey               = PolicyXPub
	PolicyExtendedPrivateKey              = PolicyXPrv
)
