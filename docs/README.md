# BSV Script Templates Documentation

Welcome to the documentation for the BSV Script Templates repository. This documentation provides detailed information about the various Bitcoin SV script templates available in this repository.

## Structure

- Templates `./template/*/README.md` - Documentation for individual script templates
- [Contributing](../.github/CONTRIBUTING.md) - Guide for contributors

## Getting Started

### Installation

```bash
go get github.com/bsv-blockchain/go-script-templates
```

### Basic Usage

1. Import the specific template you need:
   ```go
   import "github.com/bsv-blockchain/go-script-templates/template/bsocial"
   ```

2. Use the template functions to create or decode transactions:
   ```go
   // See individual template documentation for specific usage examples
   ```

## Available Templates

The repository includes templates for various use cases:

- **[BitCom](../template/bitcom/README.md)** - BitCom protocol utilities (B, MAP, AIP)
- **[BSocial](../template/bsocial)** - Social media actions using BitcoinSchema.org standards
- **[BSV20](../template/bsv20)** - BSV20 token standard implementation
- **[BSV21](../template/bsv21)** - BSV21 token standard implementation including LTM and POW20
- **[Cosign](../template/cosign)** - Co-signing transactions with multiple parties
- **[Inscription](../template/inscription)** - On-chain NFT-like inscriptions
- **[Lockup](../template/lockup)** - Time-locked transactions
- **[OrdLock](../template/ordlock)** - Locking and unlocking functionality for ordinals
- **[OrdP2PKH](../template/ordp2pkh)** - Ordinal-aware P2PKH transactions
- **[P2PKH](../template/p2pkh)** - Standard Pay-to-Public-Key-Hash transactions
- **[Shrug](../template/shrug)** - Experimental template for demo purposes

## Support

If you encounter issues or have questions, please:

1. Check the documentation for the specific template
2. Search existing GitHub issues
3. Open a new issue if needed

## License

The code in this repository is licensed under the Open BSV License. See [LICENSE](../LICENSE) for details.
