package wallet

type WalletDB interface {
	SaveWallet(*Wallet) error
	GetWallets() []Wallet
	DeleteWallet(WalletID) error
	Close()
}
