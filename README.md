# cardano-wallet

A simple Cardano wallet CLI written in Go.

## Installation from source

Clone the repository using `git clone`

```
$ git clone https://github.com/echovl/cardano-wallet.git
```

Compile the source code and install the executable

```
$ make && sudo make install
```

## Dependencies

For balance and transfer commands `cardano-node` and `cardano-cli` are required. 
You can install them using this guide https://docs.cardano.org/projects/cardano-node/en/latest/getting-started/install.html

## Getting started

First create a new wallet and generate your mnemonic squence:

```
$ cardano-wallet new-wallet myWallet -p simplePassword
mnemonic: banner capital gift plate worth sand pass canvas pave decade pig borrow cruel lunar arena
```

If you already have a wallet you can restore it using a mnemonic and password:running this command:

```
$ cardano-wallet new-wallet restoredWallet -m=talent,risk,require,split,leave,script,panel,slight,entire,soap,chase,pill,grant,laugh,fringe -p simplePassword
```

You can inspect your wallets using the `list-wallets` command:

```
$ cardano-wallet list-wallets
ID              NAME      ADDRESS
wl_uu4FmZvNYG   myWallet  1
```

By default a new wallet is created with one payment address, you can create more addresses running the following command:

```
$ cardano-wallet new-address wl_uu4FmZvNYG
New address addr_test1vz8vyz6pk6hwgwqz239rcyk52e659aefa8g08amm80tq8ag9eng6q
```

To get all addresses run:

```
$ cardano-wallet list-address wallet_WGejugqca4 --testnet
PATH                      ADDRESS
m/1852'/1815'/0'/0/0      addr_test1vpfla0wgltpjwxzt52p7wkn720eact33udlq9z8xrc6cypc3c70f5

```

You can get your balance running:

```
$ cardano-wallet balance wallet_WGejugqca4 --testnet
ASSET                     AMOUNT
Lovelace                  1000000000
```
