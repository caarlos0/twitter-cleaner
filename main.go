package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/alecthomas/kingpin"
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	_ "github.com/joho/godotenv/autoload"
)

var (
	version           = "dev"
	app               = kingpin.New("twitter-cleaner", "clean up old twitter tweets and likes")
	keeplist          = app.Flag("keep", "do not delete tweets that contain these words or tweet IDs").Strings()
	maxAge            = app.Flag("max-age", "delete tweets older than this").Default("720h").Duration()
	consumerKey       = app.Flag("twitter-consumer-key", "your twitter consumer key").Envar("TWITTER_CONSUMER_KEY").Required().String()
	consumerSecret    = app.Flag("twitter-consumer-secret", "your twitter consumer secret").Envar("TWITTER_CONSUMER_SECRET").Required().String()
	accessToken       = app.Flag("twitter-access-token", "your twitter access token").Envar("TWITTER_ACCESS_TOKEN").Required().String()
	accessTokenSecret = app.Flag("twitter-access-token-secret", "your twitter access token secret").Envar("TWITTER_ACCESS_TOKEN_SECRET").Required().String()
	archiveFolder     = app.Flag("twitter-archive-path", "path to the twitter archive, if you pass this flag, twitter-cleaner will try to delete your tweets from there too").File()
	debug             = app.Flag("debug", "enables debug logs").Bool()
)

func getTimeline(api *anaconda.TwitterApi, maxID string) ([]anaconda.Tweet, error) {
	var args = url.Values{}
	args.Add("count", "200")        // Twitter only returns most recent 20 tweets by default, so override
	args.Add("include_rts", "true") // When using count argument, RTs are excluded, so include them as recommended
	if len(maxID) > 0 {
		args.Set("max_id", maxID)
	}

	timeline, err := api.GetUserTimeline(args)
	if err != nil {
		return make([]anaconda.Tweet, 0), fmt.Errorf("failed to get timeline: %w", err)
	}
	return timeline, nil
}

func getFaves(api *anaconda.TwitterApi, maxID string) ([]anaconda.Tweet, error) {
	var args = url.Values{}
	args.Add("count", "200") // Twitter only returns most recent 20 tweets by default, so override
	if len(maxID) > 0 {
		args.Set("max_id", maxID)
	}

	faves, err := api.GetFavorites(args)
	if err != nil {
		return make([]anaconda.Tweet, 0), fmt.Errorf("failed to get favorites: %w", err)
	}
	return faves, nil
}

func isWhitelisted(id int64, text string) bool {
	tweetID := strconv.FormatInt(id, 10)
	for _, w := range *keeplist {
		if w == tweetID || strings.Contains(text, w) {
			return true
		}
	}
	return false
}

func deleteFromTimeline(api *anaconda.TwitterApi) error {
	var deletedCount int64
	var maxID string

	for i := 1; i <= 10; i++ {
		timeline, err := getTimeline(api, maxID)
		if err != nil {
			return fmt.Errorf("failed to clean up timeline: %w", err)
		}
		log.Debugf("timeline length %d", len(timeline))

		for _, t := range timeline {
			deleted, err := deleteTweet(api, t)
			if err != nil {
				return err
			}
			if deleted {
				deletedCount++
			}
			maxID = fmt.Sprintf("%d", t.Id)
		}
	}

	log.Infof("deleted %d tweets from twitter timeline", deletedCount)
	return nil
}

func deleteTweet(api *anaconda.TwitterApi, t anaconda.Tweet) (bool, error) {
	createdTime, err := t.CreatedAtTime()
	if err != nil {
		return false, fmt.Errorf("couldt not parse time '%s' from tweet '%s': %w", t.CreatedAt, t.IdStr, err)
	}

	if time.Since(createdTime) < *maxAge || isWhitelisted(t.Id, t.Text) {
		return false, nil
	}

	var derr error
	if t.Retweeted {
		log.WithFields(log.Fields{
			"id":   t.Id,
			"time": createdTime,
			"text": t.Text,
		}).Debug("unretweeting tweet")
		_, derr = api.UnRetweet(t.Id, true)
	} else if !t.Favorited {
		log.WithFields(log.Fields{
			"id":   t.Id,
			"time": createdTime,
			"text": t.Text,
		}).Debug("deleting tweet")
		_, derr = api.DeleteTweet(t.Id, true)
	}

	if derr == nil {
		return true, nil
	}

	var aerr *anaconda.ApiError
	if errors.As(derr, &aerr) {
		if aerr.StatusCode == 403 {
			log.WithError(derr).Warn("ignored 403 while deleting tweet")
			return false, nil
		}
		if aerr.StatusCode == 404 {
			log.WithError(derr).Warn("ignored 404 while deleting tweet")
			return false, nil
		}
	}
	return false, fmt.Errorf("failed to delete tweet: %w", derr)
}

func unFavorite(api *anaconda.TwitterApi) error {
	var unfavedCount int64
	var maxID string

	for i := 1; i <= 10; i++ {
		faves, err := getFaves(api, maxID)
		if err != nil {
			return fmt.Errorf("could not get favortes: %w", err)
		}
		log.Debugf("favorites length %d", len(faves))

		for _, t := range faves {
			unfaved, err := unFavoriteTweet(api, t)
			if err != nil {
				return err
			}
			if unfaved {
				unfavedCount++
			}
			maxID = fmt.Sprintf("%d", t.Id)
		}
	}

	log.Infof("unfavorited %d tweets from twitter timeline", unfavedCount)
	return nil
}

func unFavoriteTweet(api *anaconda.TwitterApi, t anaconda.Tweet) (bool, error) {
	if !t.Favorited {
		return false, nil
	}
	createdTime, err := t.CreatedAtTime()
	if err != nil {
		return false, fmt.Errorf("couldt not parse time '%s' from tweet '%s': %w", t.CreatedAt, t.IdStr, err)
	}
	if time.Since(createdTime) < *maxAge || isWhitelisted(t.Id, t.Text) {
		return false, nil
	}

	log.WithFields(log.Fields{
		"id":   t.Id,
		"time": createdTime,
		"text": t.Text,
	}).Debug("unfavoriting tweet")
	if _, err := api.Unfavorite(t.Id); err != nil {
		var aerr *anaconda.ApiError
		if errors.As(err, &aerr) {
			if aerr.StatusCode == 403 {
				log.WithError(err).Warn("ignoring 403")
				return false, nil
			}
			if aerr.StatusCode == 404 {
				log.WithError(err).Warn("ignoring 404")
				return false, nil
			}
		}
		return false, fmt.Errorf("failed to unfavorite tweet: %w", err)
	}
	return true, nil
}

func deleteFromData(api *anaconda.TwitterApi) error {
	var data = *archiveFolder
	bts, err := ioutil.ReadFile(filepath.Join(data.Name(), "data/tweet.js"))
	if err != nil {
		return err
	}

	var tweets []struct {
		Tweet struct {
			ID string `json:"id"`
		} `json:"tweet"`
	}

	if err := json.Unmarshal(bytes.TrimPrefix(bts, []byte("window.YTD.tweet.part0 = ")), &tweets); err != nil {
		return err
	}

	f, err := os.OpenFile(filepath.Join(data.Name(), "data/handled_tweets.txt"), os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	ids, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	var deletedCount int64
	for _, t := range tweets {
		if bytes.Contains(ids, []byte(t.Tweet.ID)) {
			// ignoring handled tweet
			continue
		}
		tweet, err := getTweet(api, t.Tweet.ID)
		if err != nil {
			return err
		}
		if tweet.Id == 0 { // empty tweet, 404 probably
			if _, err := f.WriteString(t.Tweet.ID + "\n"); err != nil {
				return err
			}
			continue
		}
		deleted, err := deleteTweet(api, tweet)
		if err != nil {
			return err
		}
		if deleted {
			deletedCount++
		}

		if _, err := f.WriteString(t.Tweet.ID + "\n"); err != nil {
			return err
		}
	}

	log.Infof("deleted %d tweets from twitter archive", deletedCount)
	return nil
}

func getTweet(api *anaconda.TwitterApi, s string) (anaconda.Tweet, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return anaconda.Tweet{}, err
	}
	tweet, err := api.GetTweet(id, url.Values{})
	if err != nil {
		var aerr *anaconda.ApiError
		if errors.As(err, &aerr) {
			if aerr.StatusCode == 404 {
				log.WithError(err).Warn("ignoring 404")
				return anaconda.Tweet{}, nil
			}
			if aerr.StatusCode == 403 {
				log.WithError(err).Warn("got 403, will try to delete/unfavorite anyway")
				// I'm not suspended, so its probably a RT or a fav...
				return anaconda.Tweet{
					Id:        id,
					Retweeted: true,
					Favorited: true,
					CreatedAt: "Thu Nov 13 00:00:00 +0000 2000",
				}, nil
			}
			if aerr.StatusCode == 401 {
				log.WithError(err).Warn("got 401, will try to delete/unfavorite anyway")
				// I'm haven't bloecked myself, so its probably a RT or a fav...
				return anaconda.Tweet{
					Id:        id,
					Retweeted: true,
					Favorited: true,
					CreatedAt: "Thu Nov 13 00:00:00 +0000 2000",
				}, nil
			}
		}
		return anaconda.Tweet{}, err
	}
	return tweet, nil
}

func unlikeFromData(api *anaconda.TwitterApi) error {
	var data = *archiveFolder

	bts, err := ioutil.ReadFile(filepath.Join(data.Name(), "data/like.js"))
	if err != nil {
		return err
	}

	var likes []struct {
		Like struct {
			TweetID string `json:"tweetId"`
		} `json:"like"`
	}

	if err := json.Unmarshal(bytes.TrimPrefix(bts, []byte("window.YTD.like.part0 = ")), &likes); err != nil {
		return err
	}

	f, err := os.OpenFile(filepath.Join(data.Name(), "data/handled_likes.txt"), os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	ids, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	var unfavCount int64
	for _, t := range likes {
		if bytes.Contains(ids, []byte(t.Like.TweetID)) {
			// ignoring handled tweet
			continue
		}
		tweet, err := getTweet(api, t.Like.TweetID)
		if err != nil {
			return err
		}
		if tweet.Id == 0 { // empty tweet, 404 probably
			if _, err := f.WriteString(t.Like.TweetID + "\n"); err != nil {
				return err
			}
			continue
		}
		unfaved, err := unFavoriteTweet(api, tweet)
		if err != nil {
			return err
		}
		if unfaved {
			unfavCount++
		}
		if _, err := f.WriteString(t.Like.TweetID + "\n"); err != nil {
			return err
		}
	}

	log.Infof("unfavorited %d tweets from archive", unfavCount)
	return nil
}

func main() {
	app.Author("Carlos Alexandro Becker <root@carlosbecker.dev>")
	app.Version("twitter-cleaner version " + version)
	app.VersionFlag.Short('v')
	app.HelpFlag.Short('h')
	_ = kingpin.MustParse(app.Parse(os.Args[1:]))

	log.SetHandler(cli.Default)
	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	anaconda.SetConsumerKey(*consumerKey)
	anaconda.SetConsumerSecret(*consumerSecret)
	api := anaconda.NewTwitterApi(*accessToken, *accessTokenSecret)
	api.SetLogger(anaconda.BasicLogger)

	if *archiveFolder != nil {
		log.Infof("deleting tweets from twitter archive at %s", (*archiveFolder).Name())
		if err := deleteFromData(api); err != nil {
			log.WithError(err).Fatal("failed to delete tweets from archive")
		}

		log.Infof("unfavoriting tweets from twitter archive at %s", (*archiveFolder).Name())
		if err := unlikeFromData(api); err != nil {
			log.WithError(err).Fatal("failed to unfavorite tweets from archive")
		}
	}

	log.Info("deleting tweets from twitter timeline")
	if err := deleteFromTimeline(api); err != nil {
		log.WithError(err).Fatal("failed to delete tweets from timeline")
	}

	log.Info("unfavoriting tweets from twitter timeline")
	if err := unFavorite(api); err != nil {
		log.WithError(err).Fatal("failed to unfavorite tweets from timeline")
	}

	log.Info("done")
}
