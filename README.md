# twitter-cleaner

Automatically delete tweets, retweets, and favorites from your timeline, and, if provided, from your twitter archive as well.

## Usage

You'll need [API keys](https://github.com/caarlos0/twitter-cleaner#api-keys).
Once you have them, you can either create a `.env` file or via args, you check
both the args and environment variable names using `twitter-cleaner --help`.

If you have a `.env` file, basic usage is:

```sh
twitter-cleaner
```

So, basically:

- Twitter secrets need to be provided via flags, environment variables or `.env`;
- By default, it will delete everything older than 30 days (720 hours), this can be customized via `--max-age`;
- You can prevent specific tweet IDs or tweets with specific words from being deleted by using the `--keeplist` flag;

## Advanced Usage

### Deleting from twitter archive

The twitter API only returns the last N tweets, so you can't get your whole history from it. You can, though, request your twitter data and use it to delete things.

You can request yours [here](https://twitter.com/settings/your_twitter_data). It usually takes a couple of days to arrive at your e-mail.

Once you have it, download and extract it, and then pass the resulting folder to twitter-cleaner with the `--twitter-archive-path` flag, e.g.:

```sh
twitter-cleaner --twitter-archive-path ~/Downloads/twitter-2020-12-01-asdasdasd
```

While running, twitter-cleaner will create 2 files:

- `~/Downloads/twitter-2020-12-01-asdasdasd/data/handled_tweets.txt`
- `~/Downloads/twitter-2020-12-01-asdasdasd/data/handled_likes.txt`

This is to prevent re-trying every tweet if you stop and run it again. If you want to force a full run, delete those files.

> PS: Deleting your archive will probably span across a couple of days.

### Rate limits

Once a rate limit is hit, twitter-cleaner will wait and try again. So you can basically just leave it alone and it will figure itself out.

## API Keys

To get all the keys needed, you'll need to [create a new twitter app](https://developer.twitter.com/en/portal/apps/new).

Then, go to the app's *Settings* > *Keys and tokens*. There you'll find the **API key & secret** (`--twitter-consumer-key` and `--twitter-consumer-secret`) and can generate the **Access token & secret** (`--twitter-access-token` and `--twitter-access-token-secret`).

You can pass them via flags, environment variables or via `.env` file.

## Install

**homebrew**:

```sh
brew install caarlos0/tap/twitter-cleaner
```

**docker**:

```sh
docker run --rm caarlos0/twitter-cleaner
```

**apt**:

```sh
echo 'deb [trusted=yes] https://repo.caarlos0.dev/apt/ /' | sudo tee /etc/apt/sources.list.d/caarlos0.list
sudo apt update
sudo apt install twitter-cleaner
```

**yum**:

```sh
echo '[caarlos0]
name=caarlos0
baseurl=https://repo.caarlos0.dev/yum/
enabled=1
gpgcheck=0' | sudo tee /etc/yum.repos.d/caarlos0.repo
sudo yum install twitter-cleaner
```

**deb/rpm/apk**:

Download the `.apk`, `.deb` or `.rpm` from the [releases page][releases] and install with the appropriate commands.

**manually**:

Download the pre-compiled binaries from the [releases page][releases] or clone the repo build from source.

[releases]: https://github.com/caarlos0/twitter-cleaner/releases


## Stargazers over time

[![Stargazers over time](https://starchart.cc/caarlos0/twitter-cleaner.svg)](https://starchart.cc/caarlos0/twitter-cleaner)


## Acknowledgements

This tool is heavily based on https://github.com/karan/fleets, which the main difference being that it only deletes from timeline, while this one deletes from the archive as well.
