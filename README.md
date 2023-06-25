# Carnac Swap

Carnac Swap is an experimental script for automated swap trading between virtual crypto pairs on the KuCoin exchange. This README provides instructions for setting up and using the script.

## DISCLAIMER

This script is experimental and it is used at the risk of the user. There are no guarantees of profit and no guarantees that there aren't bugs that could cause scenarios where funds could be lost due to unexpected trades. All responsibility is on the user of the script and this is for research and entertainment purposes only.

The application uses 100% of the USDT in the trading account associated with the API to make swaps, so it is recommended that the user sets up and uses a sub-account to run this application.

## Prerequisites

- A [KuCoin](https://www.kucoin.com/) account
- A KuCoin API with trade-only permissions

## Setting up KuCoin account and API

Follow these steps to set up a KuCoin account and generate an API with trade-only permissions:

1. Create a KuCoin account [here](https://www.kucoin.com/ucenter/signup).
2. Once you've logged in, navigate to your account and click on API Management.
3. Click on "Create API", set the permissions to "Trade", and complete the verification process.
4. You will be provided with an API Key, Secret Key, and a Passphrase. Keep these safe, as they will be needed to configure the script.

### Setting up a KuCoin Sub-account

Creating a sub-account allows you to separate your funds and trading strategies. Follow these steps to set up a KuCoin sub-account:

1. Log into your KuCoin account.
2. Click on your account at the top right corner and select "Sub-Account".
3. Click on "Create" and fill in the required information.
4. Once the sub-account is created, you can transfer funds into it and use it for trading.

## Trading Balance Requirement

In order for Carnac Swap to successfully execute swaps, you need to ensure that you have at least $50 worth of each coin you wish to trade within your trading account. This is a minimum requirement for the trading process to work properly. Reminder: As with any investment, do not invest more than you are willing to lose while using this application. This is highly experimental.

## Configuring the Script

The `config.json` file in the `1.02` version folder needs to be updated with your KuCoin API information, your chosen settings, and optional Telegram bot credentials.

```json
{
  "api_key": "xxxxxxx",
  "secret_key": "xxxxxxx",
  "passphrase": "xxxxxx",
  "sell_threshold": 0.4,
  "base_amount": "1%",
  "quote_amount": "1%",
  "custom_pairs": "LUNC-USDT/USTC-USDT,BTC-USDT/USTC-USDT",
  "telegram_bot_token": "xxxxxxx",
  "telegram_chat_id": "@xxxxxxx"
}
```
- Replace the `xxxxxxx` and `@xxxxxxx` values with your actual KuCoin API Key, Secret Key, Passphrase, and Telegram bot credentials.
- `sell_threshold` is the minimum percentage gain for a swap to be considered. The minimum recommended value is `0.5` (0.4%). This accounts for the 0.1% fee for each buy and sell transaction (0.2% in total) and gives a little wiggle room for slippage.
- `base_amount` and `quote_amount` are the percentages of your COIN balance (not USDT) that will be used for each swap. The recommended value is `"1%"`.
- `custom_pairs` are the trading pairs that the script will monitor for swaps. They must be in the format `BASE-QUOTE/BASE-QUOTE` and multiple pairs should be comma-delimited with no spaces after the comma.
- `telegram_bot_token` and `telegram_chat_id` are optional parameters for Telegram bot integration. If you have a Telegram bot, you can enter its token and your chat ID to receive updates from the script.

## Running the Script

After the `config.json` file has been updated, the script can be run using the `monitor` binary file.

### Windows

Open Command Prompt and navigate to the directory containing the `monitor.exe` file. You can change directories using the `cd` command:

```bash
cd path\to\directory
```

Replace path\to\directory with the path to the directory containing the monitor.exe file. Once you're in the correct directory, run the following command to start the script:

```bash
monitor.exe
```

### MacOS and Linux
Open Terminal and navigate to the directory containing the `monitor` file. You can change directories using the `cd` command.

```bash
cd path/to/directory
```

Replace `path/to/directory` with the path to the directory containing the `monitor` file. Once you're in the correct directory, you may need to give the `monitor` file execute permissions. You can do this with the `chmod` command.

```bash
chmod +x monitor
```

Then, you can run the script with the following command:

```bash
./monitor
```

## Support

Remember, this script is experimental and should be used at your own risk. If you encounter any issues or have any suggestions, please raise an issue on this GitHub repository.

## License

This project is licensed under the MIT License.
