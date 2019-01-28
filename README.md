# Image Fetcher
[![Go Report Card](https://goreportcard.com/badge/github.com/ericdaugherty/imagefetcher)](https://goreportcard.com/report/github.com/ericdaugherty/imagefetcher)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/ericdaugherty/imagefetcher/LICENSE)

ImageFetcher fetches an image from a URL, crops it, and uploads it to an S3 bucket or stores it locally (or both).

It also supports pausing it once daily (ex, at night).

```
Usage:
  -dir string
        The local directory to store the captures images in.
  -imageURL string
        The URL of the image to fetch.
  -rect string
        The x/y coordinates that define the rectangle to use to crop in the form of x,y,x,y
  -s3Bucket string
        The S3 bucket name (not the ARN) to upload the snapshots to.
  -s3Region string
        The AWS region to use. (default "us-east-1")
  -sleepHour int
        The hour to pause image capture (0-23)
  -wakeHour int
        The hour to resume image capture (0-23)```

