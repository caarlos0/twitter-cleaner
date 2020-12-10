# twitter-cleaner

Automatically delete tweets, retweets, and favorites from your timeline, and, if provided, from your twitter archive as well.

You can run it with `--help` to get the options.
Twitter secrets are also auto-loaded from a `.env` file if it exists.

By default, it will delete everything older than 30 days (720 hours), this can be customized via `--max-age`.

You can also prevent specific tweet IDs or tweets with specific words from being deleted by using the `--keeplist` flag.

## Deleting from twitter archive

You can request yours [here](https://twitter.com/settings/your_twitter_data). It usually takes a couple of days to arrive at your e-mail.

Once you have it, download and extract it, then pass the resulting folder to twitter-cleaner with the `--twitter-archive-path` flag, e.g.:

```sh
twitter-cleaner --twitter-archive-path ~/Downloads/twitter-2020-12-01-asdasdasd
```

While running, it will create 2 files:

- `~/Downloads/twitter-2020-12-01-asdasdasd/data/handled_tweets.txt`
- `~/Downloads/twitter-2020-12-01-asdasdasd/data/handled_likes.txt`

This is to prevent re-trying every tweet if you stop and run it again. If you want to force a full run, delete those files.

PS: Deleting your archive will probably span across a couple of days.

## Acknowledgements

This tool is heavily based on https://github.com/karan/fleets, which the main difference being that it only deletes from timeline, while this one deletes from the archive as well.
