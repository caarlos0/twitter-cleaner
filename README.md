# twitter-cleaner

Automatically delete tweets, retweets, and favorites from your timeline, and, if provided, from your twitter archive as well.

## Usage

- download the archive for your OS in the [releases page](https://github.com/caarlos0/twitter-cleaner/releases/latest).
- extract it
- run it with `--help` to get the full list of options

### Usage tips and details

- Twitter secrets are also auto-loaded from a `.env` file if it exists;
- By default, it will delete everything older than 30 days (720 hours), this can be customized via `--max-age`;
- You can also prevent specific tweet IDs or tweets with specific words from being deleted by using the `--keeplist` flag;

## Deleting from twitter archive

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

## Rate limits

Once a rate limit is hit, twitter-cleaner will wait and try again. So you can basically just leave it alone and it will figure itself out.

## Acknowledgements

This tool is heavily based on https://github.com/karan/fleets, which the main difference being that it only deletes from timeline, while this one deletes from the archive as well.
